package server

import (
	"net/http"

	mcppkg "github.com/cometline/cometmind/internal/mcp"
	"github.com/gin-gonic/gin"
)

func (a *App) handleListMCPServers(c *gin.Context) {
	if a.mcpMgr == nil {
		c.JSON(http.StatusOK, gin.H{"servers": []mcppkg.ServerRuntimeStatus{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"servers": a.mcpMgr.ListServers()})
}

func (a *App) handleListMCPTools(c *gin.Context) {
	if a.mcpMgr == nil {
		c.JSON(http.StatusOK, gin.H{"tools": []mcppkg.ToolInfo{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tools": a.mcpMgr.ListToolInfos()})
}

func (a *App) handleTestMCPServer(c *gin.Context) {
	if a.mcpMgr == nil {
		writeError(c, http.StatusServiceUnavailable, "mcp_unavailable", "MCP manager is not initialized")
		return
	}
	id := c.Param("id")
	result := a.mcpMgr.TestServer(c.Request.Context(), id)
	status := http.StatusOK
	if !result.OK {
		status = http.StatusBadGateway
	}
	c.JSON(status, result)
}

func (a *App) handleReconnectMCPServer(c *gin.Context) {
	if a.mcpMgr == nil {
		writeError(c, http.StatusServiceUnavailable, "mcp_unavailable", "MCP manager is not initialized")
		return
	}
	id := c.Param("id")
	if err := a.mcpMgr.Reconnect(c.Request.Context(), id); err != nil {
		writeError(c, http.StatusBadGateway, "mcp_reconnect_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
