package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestManagerInMemoryToolCall(t *testing.T) {
	ctx := context.Background()
	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "v0.0.1"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "Echo input",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{"type": "string"},
			},
		},
	}, func(_ context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
		msg, _ := args["message"].(string)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, nil, nil
	})

	t1, t2 := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer serverSession.Close()

	cfg := ServerConfig{ID: "demo", Name: "Demo", Enabled: true, Transport: TransportStdio}
	conn, err := connectServerWithTransport(ctx, cfg, t2)
	if err != nil {
		t.Fatalf("connectServerWithTransport: %v", err)
	}
	defer conn.session.Close()

	if len(conn.tools) != 1 || conn.tools[0].Name != "echo" {
		t.Fatalf("tools = %#v, want echo", conn.tools)
	}

	if ToolName(cfg.ID, conn.tools[0].Name) != "mcp_demo_echo" {
		t.Fatalf("registry name mismatch")
	}

	res, err := conn.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "echo",
		Arguments: map[string]any{"message": "hello"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if text != "hello" {
		t.Fatalf("output = %q", text)
	}
}

func TestManagerStartAndList(t *testing.T) {
	ctx := context.Background()
	mgr := NewManager(Config{
		Enabled: true,
		Servers: []ServerConfig{{ID: "other", Name: "Other", Enabled: true, Transport: TransportStdio, Command: "false"}},
	})
	mgr.Start(ctx)
	statuses := mgr.ListServers()
	if len(statuses) != 1 {
		t.Fatalf("statuses = %d, want 1", len(statuses))
	}
	if statuses[0].Status != StatusError {
		t.Fatalf("status = %q, want error", statuses[0].Status)
	}
}

func TestToolName(t *testing.T) {
	tests := []struct {
		serverID string
		toolName string
		want     string
	}{
		{"github", "create_issue", "mcp_github_create_issue"},
		{"demo", "echo", "mcp_demo_echo"},
		{"my-server", "search", "mcp_my-server_search"},
		{"ctx7", "resolve-library-id", "mcp_ctx7_resolve-library-id"},
		{"plugin", "browser/navigate", "mcp_plugin_browser_navigate"},
	}
	for _, tt := range tests {
		if got := ToolName(tt.serverID, tt.toolName); got != tt.want {
			t.Fatalf("ToolName(%q, %q) = %q, want %q", tt.serverID, tt.toolName, got, tt.want)
		}
	}
}
