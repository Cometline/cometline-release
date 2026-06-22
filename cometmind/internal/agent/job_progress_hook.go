package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cometline/cometmind/internal/jobs"
)

const defaultJobProgressNudgeAfterTools = 3

// OngoingJobLookup resolves the ongoing job assigned to a session, if any.
type OngoingJobLookup interface {
	JobForSession(ctx context.Context, sessionID string) (jobs.Job, bool, error)
}

// JobProgressTracker nudges the model to call update_job after several tool
// calls without a progress write during an ongoing job turn.
type JobProgressTracker struct {
	JobID             string
	SinceLastProgress int
	active            bool
	threshold         int
}

func newJobProgressTracker(ctx context.Context, lookup OngoingJobLookup, sessionID string) *JobProgressTracker {
	t := &JobProgressTracker{threshold: defaultJobProgressNudgeAfterTools}
	if lookup == nil || strings.TrimSpace(sessionID) == "" {
		return t
	}
	job, ok, err := lookup.JobForSession(ctx, sessionID)
	if err != nil || !ok {
		return t
	}
	t.JobID = job.ID
	t.active = true
	return t
}

// ObserveTool records a tool call. It returns true when a progress nudge should
// be injected into the next step's system prompt.
func (t *JobProgressTracker) ObserveTool(name string, input json.RawMessage) bool {
	if t == nil || !t.active || t.JobID == "" {
		return false
	}
	switch name {
	case "complete_job", "release_job":
		t.active = false
		return false
	case "update_job":
		if updateJobHasProgress(input) {
			t.SinceLastProgress = 0
		} else {
			t.SinceLastProgress++
		}
	default:
		t.SinceLastProgress++
	}
	if t.SinceLastProgress < t.threshold {
		return false
	}
	t.SinceLastProgress = 0
	return true
}

func updateJobHasProgress(input json.RawMessage) bool {
	if len(input) == 0 {
		return false
	}
	var in struct {
		Progress string `json:"progress"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return false
	}
	return strings.TrimSpace(in.Progress) != ""
}

// FormatJobProgressNudgeBlock returns a system-prompt block reminding the
// model to persist job progress.
func FormatJobProgressNudgeBlock(jobID string) string {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return ""
	}
	return fmt.Sprintf(
		"This session is working job %q. You have run several tools without updating job progress. Call `update_job` with a brief `progress` summary of what is done and what remains before continuing.",
		jobID,
	)
}
