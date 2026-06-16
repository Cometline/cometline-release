package memory

import (
	"strconv"
	"strings"
	"testing"

	cometsdk "github.com/cometline/comet-sdk"
)

func TestBuildRetrievalQuery_EmptyHistory(t *testing.T) {
	got := BuildRetrievalQuery(RetrievalQueryInput{
		Messages: []cometsdk.Message{{
			Role:    cometsdk.RoleUser,
			Content: []cometsdk.Block{cometsdk.TextBlock{Text: "hello"}},
		}},
	})
	if got != "Current message:\nhello" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildRetrievalQuery_WithRecentContext(t *testing.T) {
	msgs := []cometsdk.Message{
		{Role: cometsdk.RoleUser, Content: []cometsdk.Block{cometsdk.TextBlock{Text: "old question"}}},
		{Role: cometsdk.RoleAssistant, Content: []cometsdk.Block{cometsdk.TextBlock{Text: "old answer"}}},
		{Role: cometsdk.RoleUser, Content: []cometsdk.Block{cometsdk.TextBlock{Text: "繼續"}}},
	}
	got := BuildRetrievalQuery(RetrievalQueryInput{Messages: msgs})
	if !strings.Contains(got, "user: old question") {
		t.Fatalf("missing recent user: %q", got)
	}
	if !strings.Contains(got, "assistant: old answer") {
		t.Fatalf("missing recent assistant: %q", got)
	}
	if !strings.Contains(got, "Current message:\n繼續") {
		t.Fatalf("missing current message: %q", got)
	}
}

func TestBuildRetrievalQuery_LimitsRecentTurns(t *testing.T) {
	var msgs []cometsdk.Message
	for i := 0; i < 5; i++ {
		msgs = append(msgs,
			cometsdk.Message{Role: cometsdk.RoleUser, Content: []cometsdk.Block{cometsdk.TextBlock{Text: "u" + strconv.Itoa(i)}}},
			cometsdk.Message{Role: cometsdk.RoleAssistant, Content: []cometsdk.Block{cometsdk.TextBlock{Text: "a" + strconv.Itoa(i)}}},
		)
	}
	msgs = append(msgs, cometsdk.Message{
		Role:    cometsdk.RoleUser,
		Content: []cometsdk.Block{cometsdk.TextBlock{Text: "current"}},
	})

	got := BuildRetrievalQuery(RetrievalQueryInput{Messages: msgs})
	if strings.Contains(got, "u0") || strings.Contains(got, "a0") {
		t.Fatalf("expected oldest turns trimmed, got %q", got)
	}
	if !strings.Contains(got, "u3") || !strings.Contains(got, "a3") {
		t.Fatalf("expected recent turns kept, got %q", got)
	}
}

func TestBuildRetrievalQuery_SkipsToolResult(t *testing.T) {
	msgs := []cometsdk.Message{
		{Role: cometsdk.RoleUser, Content: []cometsdk.Block{cometsdk.TextBlock{Text: "run tool"}}},
		{Role: cometsdk.RoleAssistant, Content: []cometsdk.Block{
			cometsdk.ToolCallBlock{ID: "tc1", Name: "read_file", Input: []byte(`{}`)},
		}},
		{Role: cometsdk.RoleToolResult, Content: []cometsdk.Block{
			cometsdk.ToolResultBlock{ToolCallID: "tc1", Content: "file contents"},
		}},
		{Role: cometsdk.RoleUser, Content: []cometsdk.Block{cometsdk.TextBlock{Text: "thanks"}}},
	}
	got := BuildRetrievalQuery(RetrievalQueryInput{Messages: msgs})
	if strings.Contains(got, "file contents") {
		t.Fatalf("tool result should be skipped: %q", got)
	}
	if !strings.Contains(got, "user: run tool") {
		t.Fatalf("expected user message in context: %q", got)
	}
}

func TestBuildRetrievalQuery_TruncatesLongSnippet(t *testing.T) {
	long := strings.Repeat("x", maxCharsPerSnippet+50)
	msgs := []cometsdk.Message{
		{Role: cometsdk.RoleUser, Content: []cometsdk.Block{cometsdk.TextBlock{Text: long}}},
		{Role: cometsdk.RoleUser, Content: []cometsdk.Block{cometsdk.TextBlock{Text: "ok"}}},
	}
	got := BuildRetrievalQuery(RetrievalQueryInput{Messages: msgs})
	if !strings.Contains(got, "…") {
		t.Fatalf("expected truncation marker: %q", got)
	}
}
