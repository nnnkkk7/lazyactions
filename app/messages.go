package app

import (
	"time"

	"github.com/nnnkkk7/lazyactions/github"
)

// === Data Loading Results ===

// WorkflowsLoadedMsg is sent when workflows have been fetched from GitHub.
type WorkflowsLoadedMsg struct {
	Workflows []github.Workflow
	Err       error
}

// RunsLoadedMsg is sent when workflow runs have been fetched from GitHub.
type RunsLoadedMsg struct {
	Runs []github.Run
	Err  error
}

// JobsLoadedMsg is sent when jobs have been fetched from GitHub.
type JobsLoadedMsg struct {
	Jobs []github.Job
	Err  error
}

// LogsLoadedMsg is sent when job logs have been fetched from GitHub.
type LogsLoadedMsg struct {
	Logs string
	Err  error
}

// === Action Results ===

// RunCancelledMsg is sent when a workflow run has been cancelled.
type RunCancelledMsg struct {
	RunID int64
	Err   error
}

// RunRerunMsg is sent when a workflow run has been rerun.
type RunRerunMsg struct {
	RunID int64
	Err   error
}

// RerunFailedJobsMsg is sent when failed jobs have been rerun.
type RerunFailedJobsMsg struct {
	RunID int64
	Err   error
}

// WorkflowTriggeredMsg is sent when a workflow has been triggered.
type WorkflowTriggeredMsg struct {
	Workflow string
	Err      error
}

// === UI State ===

// FlashMsg is sent to display a temporary message to the user.
type FlashMsg struct {
	Message  string
	Duration time.Duration
}

// FlashClearMsg is sent to clear the flash message.
type FlashClearMsg struct{}

// TickMsg is sent on each polling interval.
type TickMsg struct {
	Time time.Time
}

// === Window events ===

// WindowSizeMsg is sent when the terminal window size changes.
type WindowSizeMsg struct {
	Width  int
	Height int
}
