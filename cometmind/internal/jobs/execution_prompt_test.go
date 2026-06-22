package jobs_test

import (
	"strings"
	"testing"

	"github.com/cometline/cometmind/internal/jobs"
)

func TestExecutionPromptWithoutProgress(t *testing.T) {
	t.Parallel()
	got := jobs.ExecutionPrompt(jobs.Job{
		ID:               "job-1",
		Description:      "Fix auth",
		DefinitionOfDone: "tests pass",
	})
	if !strings.Contains(got, "Please work on: Fix auth") {
		t.Fatalf("missing description: %q", got)
	}
	if !strings.Contains(got, "Definition of done: tests pass") {
		t.Fatalf("missing DoD: %q", got)
	}
	if strings.Contains(got, "Previous progress") {
		t.Fatalf("should not include progress section: %q", got)
	}
	if !strings.Contains(got, "after each meaningful milestone") {
		t.Fatalf("missing periodic update guidance: %q", got)
	}
	if !strings.Contains(got, `job_id "job-1"`) {
		t.Fatalf("missing job id: %q", got)
	}
}

func TestExecutionPromptWithProgress(t *testing.T) {
	t.Parallel()
	got := jobs.ExecutionPrompt(jobs.Job{
		ID:          "job-2",
		Description: "Ship feature",
		Progress:    "Updated middleware; tests still failing.",
	})
	if !strings.Contains(got, "Previous progress (from an earlier attempt):") {
		t.Fatalf("missing progress header: %q", got)
	}
	if !strings.Contains(got, "Updated middleware; tests still failing.") {
		t.Fatalf("missing progress body: %q", got)
	}
	if !strings.Contains(got, "Continue from here.") {
		t.Fatalf("missing continue hint: %q", got)
	}
}
