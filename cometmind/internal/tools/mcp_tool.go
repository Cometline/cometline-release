package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcppkg "github.com/cometline/cometmind/internal/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpTool struct {
	serverID    string
	toolName    string
	description string
	parameters  json.RawMessage
	session     *mcp.ClientSession
}

func mcpToolsFromManager(mgr *mcppkg.Manager) []Tool {
	if mgr == nil {
		return nil
	}
	bindings := mgr.ToolBindings()
	out := make([]Tool, 0, len(bindings))
	for _, binding := range bindings {
		out = append(out, mcpTool{
			serverID:    binding.ServerID,
			toolName:    binding.Tool.Name,
			description: binding.Tool.Description,
			parameters:  binding.Tool.Parameters,
			session:     binding.Session,
		})
	}
	return out
}

func (t mcpTool) Spec() ToolSpec {
	desc := strings.TrimSpace(t.description)
	if desc == "" {
		desc = fmt.Sprintf("MCP tool %s from server %s", t.toolName, t.serverID)
	}
	params := t.parameters
	if len(params) == 0 {
		params = json.RawMessage(`{"type":"object","properties":{}}`)
	}
	return ToolSpec{
		Name:        mcppkg.ToolName(t.serverID, t.toolName),
		Description: desc,
		Parameters:  params,
	}
}

func (t mcpTool) Execute(ctx context.Context, input json.RawMessage) (Result, error) {
	if t.session == nil {
		return Result{OK: false, Output: "MCP session not connected"}, nil
	}
	var args map[string]any
	if len(input) > 0 && string(input) != "null" {
		if err := json.Unmarshal(input, &args); err != nil {
			return Result{OK: false, Output: "invalid tool input: " + err.Error()}, nil
		}
	}
	res, err := t.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      t.toolName,
		Arguments: args,
	})
	if err != nil {
		return Result{OK: false, Output: err.Error()}, nil
	}
	output := formatMCPCallToolResult(res)
	return Result{OK: !res.IsError, Output: output}, nil
}

func formatMCPCallToolResult(res *mcp.CallToolResult) string {
	if res == nil {
		return ""
	}
	if res.StructuredContent != nil {
		if data, err := json.MarshalIndent(res.StructuredContent, "", "  "); err == nil {
			return string(data)
		}
	}
	var parts []string
	for _, content := range res.Content {
		switch c := content.(type) {
		case *mcp.TextContent:
			if c.Text != "" {
				parts = append(parts, c.Text)
			}
		default:
			if data, err := json.Marshal(c); err == nil {
				parts = append(parts, string(data))
			}
		}
	}
	return strings.Join(parts, "\n")
}
