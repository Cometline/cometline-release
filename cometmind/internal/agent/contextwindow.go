package agent

import (
	"strings"

	"github.com/cometline/cometmind/internal/config"
)

const globalContextWindowFallback = 32_000

// ResolveContextWindow returns the model context window using settings precedence:
// model override → model metadata → provider default → method default → global fallback.
func ResolveContextWindow(cfg *config.Config, providerID, modelID string) int {
	if cfg == nil {
		return globalContextWindowFallback
	}
	modelID = strings.TrimSpace(modelID)
	providerID = strings.TrimSpace(providerID)

	entry := cfg.FindProvider(providerID)
	method := ""
	if entry != nil {
		if meta, ok := entry.ModelMetadata[modelID]; ok && meta.ContextWindow > 0 {
			return meta.ContextWindow
		}
		if entry.DefaultContextWindow > 0 {
			return entry.DefaultContextWindow
		}
		method = strings.TrimSpace(entry.Method)
	}
	if method == "" {
		method = providerID
	}
	if v := methodContextWindowDefault(method); v > 0 {
		return v
	}
	return globalContextWindowFallback
}

func methodContextWindowDefault(method string) int {
	switch strings.TrimSpace(method) {
	case config.ProviderAnthropic:
		return 200_000
	case config.ProviderOpenAI:
		return 128_000
	case config.ProviderCodex:
		return 128_000
	case config.ProviderOpencodeGo:
		return 32_000
	case config.ProviderOpenAICompat:
		return 32_000
	default:
		return 0
	}
}
