package mcp

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ToolBinding connects one discovered MCP tool to its live session.
type ToolBinding struct {
	ServerID string
	Tool     DiscoveredTool
	Session  *mcp.ClientSession
}
