package app

import (
	"context"

	"github.com/nnnkkk7/lazyactions/github"
)

type mockClientState struct {
	workflows []github.Workflow
	runs      []github.Run
	jobs      []github.Job
	logs      string
	err       error
	rateLimit int
}

func newMockClient(state *mockClientState) *github.MockClient {
	if state == nil {
		state = &mockClientState{}
	}

	return &github.MockClient{
		ListWorkflowsFunc: func(ctx context.Context, repo github.Repository) ([]github.Workflow, error) {
			return state.workflows, state.err
		},
		ListRunsFunc: func(ctx context.Context, repo github.Repository, opts *github.ListRunsOpts) ([]github.Run, error) {
			return state.runs, state.err
		},
		ListJobsFunc: func(ctx context.Context, repo github.Repository, runID int64) ([]github.Job, error) {
			return state.jobs, state.err
		},
		GetJobLogsFunc: func(ctx context.Context, repo github.Repository, jobID int64) (string, error) {
			return state.logs, state.err
		},
		CancelRunFunc: func(ctx context.Context, repo github.Repository, runID int64) error {
			return state.err
		},
		RerunWorkflowFunc: func(ctx context.Context, repo github.Repository, runID int64) error {
			return state.err
		},
		RerunFailedJobsFunc: func(ctx context.Context, repo github.Repository, runID int64) error {
			return state.err
		},
		TriggerWorkflowFunc: func(ctx context.Context, repo github.Repository, workflowFile, ref string, inputs map[string]interface{}) error {
			return state.err
		},
		RateLimitRemainingFunc: func() int {
			if state.rateLimit > 0 {
				return state.rateLimit
			}
			return 5000
		},
	}
}
