package jobs

import (
	"fmt"
	"strings"
)

// ExecutionPrompt builds the agent prompt for running a claimed job.
func ExecutionPrompt(job Job) string {
	dod := strings.TrimSpace(job.DefinitionOfDone)
	if dod == "" {
		dod = "(none specified)"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Please work on: %s\n\nDefinition of done: %s\n", job.Description, dod)
	if progress := strings.TrimSpace(job.Progress); progress != "" {
		b.WriteString("\nPrevious progress (from an earlier attempt):\n")
		b.WriteString(progress)
		b.WriteString("\n\nContinue from here.\n")
	}
	fmt.Fprintf(
		&b,
		"\nWhile working, call `update_job` with `progress` after each meaningful milestone (and before long tool runs) so another session can resume if this one stops. When finished, call `complete_job` with a final progress summary.\n\n(Use job_id %q when calling job tools.)",
		job.ID,
	)
	return b.String()
}
