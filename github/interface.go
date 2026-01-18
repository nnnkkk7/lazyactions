package github

import "context"

// Client is the interface for GitHub API operations.
// It enables dependency injection for testing.
type Client interface {
	// Workflows
	ListWorkflows(ctx context.Context, repo Repository) ([]Workflow, error)

	// Runs
	ListRuns(ctx context.Context, repo Repository, opts *ListRunsOpts) ([]Run, error)
	CancelRun(ctx context.Context, repo Repository, runID int64) error
	RerunWorkflow(ctx context.Context, repo Repository, runID int64) error
	RerunFailedJobs(ctx context.Context, repo Repository, runID int64) error
	TriggerWorkflow(ctx context.Context, repo Repository, workflowFile, ref string, inputs map[string]interface{}) error

	// Jobs
	ListJobs(ctx context.Context, repo Repository, runID int64) ([]Job, error)

	// Logs
	GetJobLogs(ctx context.Context, repo Repository, jobID int64) (string, error)

	// Rate limiting
	RateLimitRemaining() int
}
