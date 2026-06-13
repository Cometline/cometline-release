package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	cometsdk "github.com/cometline/comet-sdk"
	"github.com/cometline/comet-sdk/llm"
	"github.com/cometline/cometmind/internal/event"
	"github.com/cometline/cometmind/internal/session"
	"github.com/cometline/cometmind/internal/tools"
)

// TurnStore is the narrow persistence seam the agent loop drives. It is the
// subset of session.Service the Runner actually needs, declared here on the
// consumer side so the loop can be unit-tested with an in-memory fake instead
// of a live SQLite database. *session.Service satisfies it.
type TurnStore interface {
	BuildSDKMessages(ctx context.Context, sessionID string) ([]cometsdk.Message, error)
	SaveTokenUsage(ctx context.Context, sessionID string, u cometsdk.TokenUsage) error
	AppendAssistantStep(ctx context.Context, sessionID, text string, reasoningBlocks []cometsdk.Block, toolCalls []cometsdk.ToolCallBlock) (session.Message, map[string]string, error)
	UpdateToolCallResult(ctx context.Context, toolCallID, result string, durMs int64, exit *int64) error
	AppendToolResultMessage(ctx context.Context, sessionID, toolCallID, output string, isErr bool) (session.Message, error)
}

// Runner executes the persisted agent loop for one user turn (which may span many tool steps).
type Runner struct {
	Provider cometsdk.Provider
	Sessions TurnStore
	Registry *tools.Registry

	MaxSteps     int
	MaxTokens    int
	SystemPrompt string
}

// Run streams CometMind-native events on ch until the turn completes or ctx is cancelled.
// The caller must receive until the channel closes.
func (r *Runner) Run(ctx context.Context, turn session.AgentTurn, ch chan<- event.Event) error {
	defer func() {
		ch <- event.Done()
	}()

	if r.MaxSteps <= 0 {
		r.MaxSteps = 50
	}
	if r.MaxTokens <= 0 {
		r.MaxTokens = 8192
	}

	steps := 0
	for steps < r.MaxSteps {
		msgs, err := r.Sessions.BuildSDKMessages(ctx, turn.ID)
		if err != nil {
			ch <- event.Errorf(err.Error(), "history")
			return err
		}

		req := BuildRequest(turn.ModelID, r.SystemPrompt, msgs, r.Registry.CometSDK(), r.MaxTokens)
		stream := llm.StreamMessage(ctx, r.Provider, req)

		for ev := range stream.Events() {
			switch e := ev.(type) {
			case cometsdk.TextDeltaEvent:
				ch <- event.TextDelta(e.Text)
			case cometsdk.ReasoningStartEvent:
				ch <- event.ReasoningStart()
			case cometsdk.ReasoningContentEvent:
				ch <- event.ReasoningDelta(e.Text)
			case cometsdk.ToolCallDoneEvent:
				ch <- event.ToolCall(e.ID, e.Name, []byte(e.Input))
			case cometsdk.StepFinishEvent:
				ch <- event.StepFinish(e.Usage)
			}
		}

		result, err := stream.Result()
		if err != nil {
			ch <- event.Errorf(err.Error(), "llm")
			return err
		}

		if err := r.Sessions.SaveTokenUsage(ctx, turn.ID, result.Usage); err != nil {
			ch <- event.Errorf(err.Error(), "db")
			return err
		}

		text := assistantPlainText(result.Message)
		reasoningBlocks := result.Message.ReasoningContent
		_, persistedToolIDs, err := r.Sessions.AppendAssistantStep(ctx, turn.ID, text, reasoningBlocks, result.ToolCalls)
		if err != nil {
			ch <- event.Errorf(err.Error(), "db")
			return err
		}

		switch result.FinishReason {
		case cometsdk.FinishStop, cometsdk.FinishMaxTokens:
			return nil
		}
		if len(result.ToolCalls) == 0 {
			return nil
		}

		for _, tc := range result.ToolCalls {
			persistedID := persistedToolIDs[tc.ID]
			if persistedID == "" {
				ch <- event.Errorf("missing persisted tool call id", "db")
				return fmt.Errorf("missing persisted tool call id for %s", tc.ID)
			}
			start := time.Now()
			res, execErr := r.Registry.Execute(ctx, tc.Name, tc.Input)
			dur := time.Since(start).Milliseconds()

			out := res.Output
			isErr := !res.OK
			if execErr != nil {
				isErr = true
				out = fmt.Sprintf("%s\n(execute error: %v)", out, execErr)
			}

			exit := int64PtrFromIntPtr(res.ExitCode)
			if err := r.Sessions.UpdateToolCallResult(ctx, persistedID, out, dur, exit); err != nil {
				ch <- event.Errorf(err.Error(), "db")
				return err
			}
			if _, err := r.Sessions.AppendToolResultMessage(ctx, turn.ID, persistedID, out, isErr); err != nil {
				ch <- event.Errorf(err.Error(), "db")
				return err
			}

			toolErr := ""
			if isErr {
				toolErr = out
			}
			ch <- event.ToolResult(tc.ID, tc.Name, out, toolErr)
		}

		steps++
	}

	ch <- event.Errorf("max steps exceeded", "max_steps")
	return fmt.Errorf("max steps exceeded")
}

func int64PtrFromIntPtr(v *int) *int64 {
	if v == nil {
		return nil
	}
	x := int64(*v)
	return &x
}

func assistantPlainText(m cometsdk.Message) string {
	var b strings.Builder
	for _, bl := range m.Content {
		if tb, ok := bl.(cometsdk.TextBlock); ok {
			b.WriteString(tb.Text)
		}
	}
	return b.String()
}
