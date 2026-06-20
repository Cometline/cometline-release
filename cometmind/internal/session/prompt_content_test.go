package session

import "testing"

func TestTruncateToolResultForPrompt(t *testing.T) {
	short := "hello"
	if got := TruncateToolResultForPrompt(short, 100); got != short {
		t.Fatalf("short content changed: %q", got)
	}
	long := string(make([]rune, 5000))
	got := TruncateToolResultForPrompt(long, MaxToolResultPromptRunes)
	if len([]rune(got)) <= MaxToolResultPromptRunes {
		t.Fatalf("expected truncation footer beyond max runes, got len=%d", len([]rune(got)))
	}
	if !contains(got, "truncated for context") {
		t.Fatalf("expected truncation marker, got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexSubstring(s, sub) >= 0)
}

func indexSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
