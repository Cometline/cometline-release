package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cometline/cometmind/internal/agent"
	"github.com/cometline/cometmind/internal/config"
	"github.com/cometline/cometmind/internal/session"
)

// Deps are shared wiring for the TUI (opened once from cmd/tui).
type Deps struct {
	Config        *config.Config
	Sessions      *session.Service
	WorkspacePath string
	WorkspaceID   string

	// NewRunner builds a runner for the currently-opened session. The runtime
	// owns provider construction so the TUI stays focused on UI concerns.
	NewRunner func(turn session.AgentTurn) (*agent.Runner, error)

	// Program is set by cmd before Run(); stream goroutines send Msgs here.
	Program *tea.Program
}
