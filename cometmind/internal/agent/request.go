package agent

import (
	"fmt"
	"strings"

	cometsdk "github.com/cometline/comet-sdk"
)

const (
	DefaultSystemPrompt = `You are CometMind, a careful coding agent working inside a single workspace on the user's machine.
You may use the provided tools to read, modify, and explore files, and to run shell commands when useful.
Prefer glob and grep for finding files and searching contents instead of run_command with find or grep.
Prefer small, verified steps. Summarize important changes clearly.`

	// maxOutputTruncationContinuations caps how many extra model steps we take
	// when a step hits the output token limit without tool calls.
	maxOutputTruncationContinuations = 2
)

// FormatOutputBudgetPromptBlock reminds the model of the per-step output cap.
func FormatOutputBudgetPromptBlock(maxTokens int) string {
	if maxTokens <= 0 {
		return ""
	}
	return fmt.Sprintf(
		"Each assistant step is capped at roughly %d output tokens. After tool results arrive, respond concisely—summarize findings and next steps instead of repeating large tool output or full file contents.",
		maxTokens,
	)
}

// FormatOutputTruncationContinueBlock nudges the model to finish after a truncated step.
func FormatOutputTruncationContinueBlock() string {
	return "Your previous assistant message in this turn was cut off at the output token limit. Continue from where you stopped. Do not repeat text already written. Be concise and finish the thought, or give a brief closing summary if the work is complete."
}

// BuildRequest constructs the outbound LLM request from history and runtime settings.
func BuildRequest(model string, system string, messages []cometsdk.Message, tools []cometsdk.Tool, maxTokens int) *cometsdk.Request {
	req := &cometsdk.Request{
		Model:     model,
		System:    system,
		Messages:  messages,
		Tools:     tools,
		MaxTokens: maxTokens,
	}
	if strings.TrimSpace(req.System) == "" {
		req.System = DefaultSystemPrompt
	}
	return req
}
