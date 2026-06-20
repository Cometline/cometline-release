package agent

import "testing"

func TestEstimatePromptTokensIncludesSummary(t *testing.T) {
	base := EstimatePromptTokens(PromptBudgetInput{
		System: "system",
	})
	withSummary := EstimatePromptTokens(PromptBudgetInput{
		System:  "system",
		Summary: "prior goals and decisions",
	})
	if withSummary <= base {
		t.Fatalf("summary should increase estimate: base=%d with=%d", base, withSummary)
	}
}
