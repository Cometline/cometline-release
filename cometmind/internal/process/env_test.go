package process

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAugmentedPathAddsUserToolDirs(t *testing.T) {
	t.Setenv("HOME", "/Users/example")

	path := AugmentedPath("/usr/bin:/bin")
	entries := filepath.SplitList(path)

	if entries[0] != filepath.Join("/Users/example", ".cometmind", "bin") {
		t.Fatalf("first PATH entry = %q, want cometmind bin", entries[0])
	}
	if !containsEntry(entries, filepath.Join("/Users/example", ".opencode", "bin")) {
		t.Fatalf("PATH %q missing ~/.opencode/bin", path)
	}
	if !containsEntry(entries, filepath.Join("/Users/example", ".local", "bin")) {
		t.Fatalf("PATH %q missing ~/.local/bin", path)
	}
	if !containsEntry(entries, "/usr/bin") || !containsEntry(entries, "/bin") {
		t.Fatalf("PATH %q missing original entries", path)
	}
}

func TestEnvReplacesPath(t *testing.T) {
	t.Setenv("PATH", "/usr/bin:/bin")

	env := Env()
	var path string
	for _, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			path = strings.TrimPrefix(kv, "PATH=")
		}
	}

	if path == "" {
		t.Fatal("PATH not set")
	}
	if !containsEntry(filepath.SplitList(path), filepath.Join(os.Getenv("HOME"), ".opencode", "bin")) {
		t.Fatalf("PATH %q missing opencode bin", path)
	}
}

func TestResolveCommandUsesAugmentedPath(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	bin := filepath.Join(home, ".cometmind", "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(bin, "cometmind")
	if err := os.WriteFile(want, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("PATH", "/usr/bin:/bin")

	got, err := ResolveCommand("cometmind")
	if err != nil {
		t.Fatalf("ResolveCommand: %v", err)
	}
	if got != want {
		t.Fatalf("ResolveCommand = %q, want %q", got, want)
	}
}

func containsEntry(entries []string, want string) bool {
	for _, entry := range entries {
		if entry == want {
			return true
		}
	}
	return false
}
