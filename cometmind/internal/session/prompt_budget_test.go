package session

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cometline/cometmind/internal/db"
)

func TestRecentWindowStartForBudget_shrinksHugeTurn(t *testing.T) {
	huge := strings.Repeat("x", MaxToolResultPromptRunes*20)
	payload, err := json.Marshal(toolResultPayload{ToolCallID: "tc1", Content: huge})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	rows := []db.Message{
		{ID: "u1", Role: "user", Content: "first question"},
		{ID: "a1", Role: "assistant", Content: ""},
		{ID: "t1", Role: "tool_result", Content: string(payload)},
		{ID: "u2", Role: "user", Content: "second question"},
	}
	calls := map[string][]db.ToolCall{
		"a1": {{ID: "tc1", MessageID: "a1", ToolName: "grep", Arguments: `{}`}},
	}

	turnOnly := RecentWindowStartForBudget(rows, calls, 10, 32_000, 2048)
	if turnOnly != 0 {
		t.Fatalf("turn-only recent start = %d, want 0", turnOnly)
	}

	budgeted := RecentWindowStartForBudget(rows, calls, 10, 4_000, 2048)
	if budgeted == 0 {
		t.Fatalf("expected budgeted recent window to drop first turn, got start=0")
	}
	if rows[budgeted].ID != "u2" {
		t.Fatalf("budgeted recent start = %d (%q), want u2", budgeted, rows[budgeted].ID)
	}
}
