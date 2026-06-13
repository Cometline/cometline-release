package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cometline/cometmind/internal/agent"
	"github.com/cometline/cometmind/internal/runtime"
	"github.com/cometline/cometmind/internal/session"
	"github.com/cometline/cometmind/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive Bubble Tea UI for CometMind sessions",
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	rt, err := runtime.New(ctx)
	if err != nil {
		return err
	}
	defer rt.Close()

	ws, err := rt.WorkspaceForCommand(ctx, "")
	if err != nil {
		return err
	}

	deps := &tui.Deps{
		Config:        rt.Config,
		Sessions:      rt.Sessions,
		WorkspacePath: ws.Path,
		WorkspaceID:   ws.ID,
		NewRunner: func(turn session.AgentTurn) (*agent.Runner, error) {
			sess := session.Session{
				ID:          turn.ID,
				WorkspaceID: ws.ID,
				ModelID:     turn.ModelID,
				ProviderID:  turn.ProviderID,
			}
			return rt.RunnerFor(sess, ws.Path)
		},
	}

	app := tui.NewApp(deps)
	p := tea.NewProgram(app, tea.WithAltScreen())
	app.SetProgram(p)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
