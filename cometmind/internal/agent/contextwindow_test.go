package agent

import (
	"testing"

	"github.com/cometline/cometmind/internal/config"
)

func TestResolveContextWindowPrecedence(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.ProviderEntry{{
			ID:                   "anthropic",
			Method:               config.ProviderAnthropic,
			DefaultContextWindow: 100_000,
			ModelMetadata: map[string]config.ModelMetadata{
				"claude-sonnet": {ContextWindow: 200_000},
			},
		}},
	}

	if got := ResolveContextWindow(cfg, "anthropic", "claude-sonnet"); got != 200_000 {
		t.Fatalf("model metadata = %d, want 200000", got)
	}
	if got := ResolveContextWindow(cfg, "anthropic", "unknown-model"); got != 100_000 {
		t.Fatalf("provider default = %d, want 100000", got)
	}
}

func TestResolveContextWindowMethodFallback(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.ProviderEntry{{
			ID:     "openai",
			Method: config.ProviderOpenAI,
		}},
	}
	if got := ResolveContextWindow(cfg, "openai", "gpt-4"); got != 128_000 {
		t.Fatalf("openai method default = %d, want 128000", got)
	}
}

func TestResolveContextWindowGlobalFallback(t *testing.T) {
	if got := ResolveContextWindow(nil, "custom", "m"); got != globalContextWindowFallback {
		t.Fatalf("global fallback = %d, want %d", got, globalContextWindowFallback)
	}
}

func TestShouldCompact(t *testing.T) {
	if ShouldCompact(1000, 10_000, 2048) {
		t.Fatal("small prompt should not compact")
	}
	if !ShouldCompact(8000, 10_000, 2048) {
		t.Fatal("large prompt should compact")
	}
}

func TestEstimateTokens(t *testing.T) {
	if got := EstimateTokens("hello world"); got < 1 {
		t.Fatalf("EstimateTokens() = %d, want >= 1", got)
	}
}
