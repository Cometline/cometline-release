package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cometline/cometmind/internal/session"
)

func TestRuntimeWiresServiceAndRunner(t *testing.T) {
	ctx := context.Background()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	rt, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rt.Close()

	if rt.Config == nil {
		t.Fatal("runtime config is nil")
	}
	if rt.Sessions == nil {
		t.Fatal("runtime sessions is nil")
	}

	// A config file should have been written to the temp home.
	if _, err := os.Stat(filepath.Join(home, ".cometmind", "config.toml")); err != nil {
		t.Fatalf("expected default config file: %v", err)
	}

	ws, err := rt.WorkspaceForCommand(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("WorkspaceForCommand() error = %v", err)
	}
	sess, err := rt.Sessions.NewSession(ctx, ws.ID, rt.Config.Model, rt.Config.Provider)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	// RunnerFor is the wiring that historically each command duplicated.
	runner, err := rt.RunnerFor(sess, ws.Path)
	if err != nil {
		t.Fatalf("RunnerFor() error = %v", err)
	}
	if runner == nil {
		t.Fatal("RunnerFor() returned nil")
	}
}

func TestRuntimeProviderForSessionUsesSessionIdentifiers(t *testing.T) {
	ctx := context.Background()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	rt, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rt.Close()

	// ProviderForSession should copy session model/provider into the config
	// passed to the provider factory, without mutating rt.Config.
	origModel := rt.Config.Model
	sess := session.Session{ModelID: "other-model", ProviderID: "other-provider"}

	_, err = rt.ProviderForSession(sess)
	if err == nil {
		t.Fatal("ProviderForSession() expected error for unknown provider, got nil")
	}
	if rt.Config.Model != origModel {
		t.Fatalf("runtime config mutated: model = %q, want %q", rt.Config.Model, origModel)
	}
}
