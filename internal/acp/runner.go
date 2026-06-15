package acp

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	acpsdk "github.com/coder/acp-go-sdk"
)

// Config controls how CometMind spawns an external ACP coding agent.
type Config struct {
	Command string
	Args    []string
	Timeout time.Duration
}

// DefaultConfig returns defaults for OpenCode in ACP mode.
func DefaultConfig() Config {
	return Config{
		Command: "opencode",
		Args:    []string{"acp"},
		Timeout: 30 * time.Minute,
	}
}

// TaskRequest is one delegated coding turn.
type TaskRequest struct {
	WorkspaceRoot string
	Task          string
	Context       string
	VerifyCommand string
	OnProgress    func(ProgressUpdate)
}

// TaskResult summarizes a delegated coding turn.
type TaskResult struct {
	Status       string
	Summary      string
	VerifyOutput string
	AgentName    string
}

// AgentRunner connects to an ACP agent subprocess and runs one prompt turn.
type AgentRunner struct {
	Config Config
	// ProcessStarter spawns the agent; defaults to exec.Command when nil.
	ProcessStarter func(ctx context.Context, cfg Config) (io.WriteCloser, io.ReadCloser, io.Closer, error)
}

// Run executes a single delegated task against an ACP agent.
func (r *AgentRunner) Run(ctx context.Context, req TaskRequest) (TaskResult, error) {
	cfg := r.Config
	if cfg.Command == "" {
		cfg = DefaultConfig()
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultConfig().Timeout
	}
	runCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	start := r.ProcessStarter
	if start == nil {
		start = defaultProcessStarter
	}
	stdin, stdout, closer, err := start(runCtx, cfg)
	if err != nil {
		return TaskResult{Status: "failed", Summary: err.Error()}, err
	}
	defer closer.Close()

	client := &WorkspaceClient{
		WorkspaceRoot: req.WorkspaceRoot,
		OnProgress:    req.OnProgress,
	}
	conn := acpsdk.NewClientSideConnection(client, stdin, stdout)

	initResp, err := conn.Initialize(runCtx, acpsdk.InitializeRequest{
		ProtocolVersion: acpsdk.ProtocolVersionNumber,
		ClientCapabilities: acpsdk.ClientCapabilities{
			Fs: acpsdk.FileSystemCapabilities{
				ReadTextFile:  true,
				WriteTextFile: true,
			},
			Terminal: true,
		},
	})
	if err != nil {
		return TaskResult{Status: "failed", Summary: err.Error()}, err
	}
	agentName := "acp-agent"
	if initResp.AgentInfo != nil && initResp.AgentInfo.Name != "" {
		agentName = initResp.AgentInfo.Name
	}

	sess, err := conn.NewSession(runCtx, acpsdk.NewSessionRequest{
		Cwd:        req.WorkspaceRoot,
		McpServers: []acpsdk.McpServer{},
	})
	if err != nil {
		return TaskResult{Status: "failed", Summary: err.Error()}, err
	}

	promptText := req.Task
	if strings.TrimSpace(req.Context) != "" {
		promptText = req.Context + "\n\nTask:\n" + req.Task
	}

	var chunks []string
	prev := client.OnProgress
	client.OnProgress = func(u ProgressUpdate) {
		if prev != nil {
			prev(u)
		}
		if u.Content != "" {
			chunks = append(chunks, u.Content)
		}
	}

	promptResp, err := conn.Prompt(runCtx, acpsdk.PromptRequest{
		SessionId: sess.SessionId,
		Prompt:    []acpsdk.ContentBlock{acpsdk.TextBlock(promptText)},
	})
	if err != nil {
		return TaskResult{Status: "failed", Summary: err.Error(), AgentName: agentName}, err
	}

	verifyOut := ""
	if strings.TrimSpace(req.VerifyCommand) != "" {
		verifyOut, _ = runVerifyCommand(runCtx, req.WorkspaceRoot, req.VerifyCommand)
	}

	status := "completed"
	if promptResp.StopReason == acpsdk.StopReasonCancelled {
		status = "cancelled"
	}

	summary := strings.TrimSpace(strings.Join(chunks, "\n"))
	if summary == "" {
		summary = fmt.Sprintf("delegation finished (%s)", promptResp.StopReason)
	}
	if verifyOut != "" {
		summary += "\n\nVerify output:\n" + verifyOut
	}

	return TaskResult{
		Status:       status,
		Summary:      summary,
		VerifyOutput: verifyOut,
		AgentName:    agentName,
	}, nil
}

// Cancel sends session/cancel when a connection is active.
func Cancel(conn *acpsdk.ClientSideConnection, sessionID acpsdk.SessionId) error {
	if conn == nil {
		return nil
	}
	return conn.Cancel(context.Background(), acpsdk.CancelNotification{SessionId: sessionID})
}

func defaultProcessStarter(ctx context.Context, cfg Config) (io.WriteCloser, io.ReadCloser, io.Closer, error) {
	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)
	cmd.Dir = "."
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, nil, nil, err
	}
	return stdin, stdout, &cmdWaitCloser{cmd: cmd}, nil
}

type cmdWaitCloser struct {
	cmd *exec.Cmd
	once sync.Once
}

func (c *cmdWaitCloser) Close() error {
	var err error
	c.once.Do(func() {
		if c.cmd.Process != nil {
			_ = c.cmd.Process.Kill()
		}
		err = c.cmd.Wait()
	})
	return err
}

func runVerifyCommand(ctx context.Context, workspaceRoot, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command) //nolint:gosec // delegated verify step
	cmd.Dir = workspaceRoot
	out, err := cmd.CombinedOutput()
	text := string(out)
	if err != nil {
		return text, err
	}
	return text, nil
}
