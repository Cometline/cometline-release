package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLabelForModel(t *testing.T) {
	t.Parallel()

	if got := LabelForModel("claude-opus-4-5"); got != "CLAUDE-OPUS-4-5" {
		t.Fatalf("LabelForModel() = %q, want CLAUDE-OPUS-4-5", got)
	}
	if got := LabelForModel("gpt_4o"); got != "GPT 4O" {
		t.Fatalf("LabelForModel() = %q, want GPT 4O", got)
	}
}

func TestModelEntriesFromSettings(t *testing.T) {
	t.Parallel()

	raw := cometlineSettingsJSON{
		Providers: []cometlineProviderJSON{
			{
				ID:            "anthropic",
				Enabled:       true,
				EnabledModels: []string{"claude-sonnet-4-5", "text-embedding-3-small"},
			},
			{
				ID:            "openai",
				Enabled:       true,
				EnabledModels: []string{"gpt-4o"},
			},
			{
				ID:            "disabled",
				Enabled:       false,
				EnabledModels: []string{"ignored"},
			},
		},
	}

	got := modelEntriesFromSettings(raw)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ProviderID != "anthropic" || got[0].ModelID != "claude-sonnet-4-5" {
		t.Fatalf("first entry = %+v", got[0])
	}
	if got[1].ProviderID != "openai" || got[1].ModelID != "gpt-4o" {
		t.Fatalf("second entry = %+v", got[1])
	}
}

func TestUpdateDefaultModel(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".cometmind")
	if err := os.MkdirAll(settingsDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(settingsDir, "cometline-settings.json")
	data, err := os.ReadFile("testdata/cometline-settings.json")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("HOME", dir)

	if err := UpdateDefaultModel("anthropic", "claude-sonnet-4-5"); err != nil {
		t.Fatalf("UpdateDefaultModel() error = %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(updated) error = %v", err)
	}
	if !containsAll(string(updated), `"defaultProviderId": "anthropic"`, `"defaultModelId": "claude-sonnet-4-5"`) {
		t.Fatalf("updated settings missing defaults: %s", string(updated))
	}
}

func containsAll(s string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
