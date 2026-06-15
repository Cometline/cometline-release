package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cometline/cometmind/internal/acp"
	"github.com/cometline/cometmind/internal/event"
	"github.com/cometline/cometmind/internal/session"
)

// DelegateCodingTask hands coding work to an external ACP agent such as OpenCode.
type DelegateCodingTask struct {
	Workspace Workspace
	Sessions  *session.Service
	ACP       acp.Config
	Runner    *acp.AgentRunner
}

func (DelegateCodingTask) Spec() ToolSpec {
	return ToolSpec{
		Name: "delegate_coding_task",
		Description: "Delegate a coding task to an external OpenCode agent over ACP. " +
			"Use for multi-file edits, refactors, and test runs. Returns a summary with verify output.",
		Parameters: json.RawMessage(`{
			"type":"object",
			"properties":{
				"task":{"type":"string","description":"Coding task for the subagent"},
				"context":{"type":"string","description":"Optional extra context, file notes, or constraints"},
				"verify_command":{"type":"string","description":"Optional shell command to run after coding, e.g. cd cometmind && go test ./..."}
			},
			"required":["task"]
		}`),
	}
}

func (d DelegateCodingTask) Execute(ctx context.Context, input json.RawMessage) (Result, error) {
	var in struct {
		Task           string `json:"task"`
		Context        string `json:"context"`
		VerifyCommand  string `json:"verify_command"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return Result{}, err
	}
	task := strings.TrimSpace(in.Task)
	if task == "" {
		return Result{OK: false, Output: "task is required"}, nil
	}
	if d.Sessions == nil {
		return Result{OK: false, Output: "delegation is not configured"}, nil
	}

	parentID := ToolSessionFrom(ctx)
	if parentID == "" {
		return Result{OK: false, Output: "missing parent session context"}, nil
	}

	parent, err := d.Sessions.GetSession(ctx, parentID)
	if err != nil {
		return Result{OK: false, Output: err.Error()}, nil
	}

	child, err := d.Sessions.NewChildSession(ctx, parent, task)
	if err != nil {
		return Result{OK: false, Output: err.Error()}, nil
	}

	emit := ProgressFrom(ctx)
	if emit != nil {
		emit(event.SubagentStarted(child.ID, task, d.ACP.Command))
	}

	_ = d.Sessions.UpdateDelegation(ctx, child.ID, "running", "")

	runner := d.Runner
	if runner == nil {
		runner = &acp.AgentRunner{Config: d.ACP}
	}
	result, runErr := runner.Run(ctx, acp.TaskRequest{
		WorkspaceRoot: d.Workspace.Root,
		Task:          task,
		Context:       in.Context,
		VerifyCommand: in.VerifyCommand,
		OnProgress: func(u acp.ProgressUpdate) {
			if emit == nil {
				return
			}
			text := u.Content
			if text == "" && u.Title != "" {
				text = u.Title
				if u.Status != "" {
					text += " (" + u.Status + ")"
				}
			}
			emit(event.SubagentProgress(child.ID, u.Kind, text))
		},
	})

	status := result.Status
	if status == "" {
		status = "failed"
	}
	if runErr != nil && status == "completed" {
		status = "failed"
	}

	summary := result.Summary
	if summary == "" && runErr != nil {
		summary = runErr.Error()
	}
	_ = d.Sessions.UpdateDelegation(ctx, child.ID, status, summary)

	if emit != nil {
		emit(event.SubagentFinished(child.ID, status, summary))
	}

	out := fmt.Sprintf("child_session_id: %s\nstatus: %s\nagent: %s\n\n%s",
		child.ID, status, result.AgentName, summary)
	ok := status == "completed"
	if runErr != nil {
		ok = false
	}
	return Result{OK: ok, Output: out}, nil
}
