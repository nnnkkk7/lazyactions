package github

import "time"

// Repository represents a GitHub repository.
type Repository struct {
	Owner string
	Name  string
}

// FullName returns the full repository name in the format "owner/name".
func (r Repository) FullName() string {
	return r.Owner + "/" + r.Name
}

// Workflow represents a GitHub Actions workflow definition.
type Workflow struct {
	ID    int64
	Name  string
	Path  string // .github/workflows/ci.yml
	State string // active, disabled
}

// Run represents a workflow run.
type Run struct {
	ID         int64
	Name       string
	Status     string // queued, in_progress, completed
	Conclusion string // success, failure, cancelled
	Branch     string
	Event      string // push, pull_request, workflow_dispatch
	CreatedAt  time.Time
	Actor      string
	URL        string
}

// IsRunning returns true if the run is in progress or queued.
func (r Run) IsRunning() bool {
	return r.Status == "in_progress" || r.Status == "queued"
}

// IsFailed returns true if the run has failed.
func (r Run) IsFailed() bool {
	return r.Conclusion == "failure"
}

// Job represents a job within a workflow run.
type Job struct {
	ID         int64
	Name       string
	Status     string // queued, in_progress, completed
	Conclusion string // success, failure, cancelled
	Steps      []Step
}

// Step represents a step within a job.
type Step struct {
	Name       string
	Status     string // queued, in_progress, completed
	Conclusion string // success, failure, skipped
	Number     int
}

// IsFailed returns true if the step has failed.
func (s Step) IsFailed() bool {
	return s.Conclusion == "failure"
}

// ListRunsOpts represents options for listing workflow runs.
type ListRunsOpts struct {
	WorkflowID int64
	Branch     string
	Event      string
	Status     string
	PerPage    int
}
