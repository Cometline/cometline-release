package memory

import (
	"strings"

	cometsdk "github.com/cometline/comet-sdk"
)

const (
	maxCharsPerSnippet      = 300
	maxRecentTurns          = 3
	maxRetrievalQueryChars  = 6000
)

// RetrievalQueryInput carries session context for auto-retrieve query expansion.
type RetrievalQueryInput struct {
	Messages []cometsdk.Message
}

// BuildRetrievalQuery assembles an embedding query from recent dialogue and the
// current user message. No LLM call is involved.
func BuildRetrievalQuery(in RetrievalQueryInput) string {
	current := lastUserMessageText(in.Messages)
	if current == "" {
		return ""
	}

	var b strings.Builder
	recent := recentContextSnippets(in.Messages)
	if len(recent) > 0 {
		b.WriteString("Recent context:\n")
		for _, line := range recent {
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("Current message:\n")
	b.WriteString(current)

	out := b.String()
	if len(out) > maxRetrievalQueryChars {
		out = out[len(out)-maxRetrievalQueryChars:]
	}
	return out
}

func recentContextSnippets(msgs []cometsdk.Message) []string {
	if len(msgs) <= 1 {
		return nil
	}
	// Exclude the final user message (current turn).
	prior := msgs[:len(msgs)-1]

	var snippets []string
	turns := 0
	for i := len(prior) - 1; i >= 0 && turns < maxRecentTurns; i-- {
		m := prior[i]
		switch m.Role {
		case cometsdk.RoleUser:
			text := truncateSnippet(messageTextFromSDK(m))
			if text == "" {
				continue
			}
			snippets = append([]string{"user: " + text}, snippets...)
			turns++
		case cometsdk.RoleAssistant:
			text := truncateSnippet(assistantTextFromSDK(m))
			if text == "" {
				continue
			}
			snippets = append([]string{"assistant: " + text}, snippets...)
		default:
			continue
		}
	}
	return snippets
}

func lastUserMessageText(msgs []cometsdk.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role != cometsdk.RoleUser {
			continue
		}
		return messageTextFromSDK(msgs[i])
	}
	return ""
}

func messageTextFromSDK(m cometsdk.Message) string {
	var b strings.Builder
	for _, bl := range m.Content {
		if tb, ok := bl.(cometsdk.TextBlock); ok {
			b.WriteString(tb.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

func assistantTextFromSDK(m cometsdk.Message) string {
	var b strings.Builder
	for _, bl := range m.Content {
		if tb, ok := bl.(cometsdk.TextBlock); ok {
			b.WriteString(tb.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

func truncateSnippet(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxCharsPerSnippet {
		return s
	}
	return s[:maxCharsPerSnippet] + "…"
}
