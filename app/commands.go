package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/lazyactions/github"
)

// fetchWorkflows creates a command to fetch workflows.
// It captures the client and repo to avoid race conditions.
// Retries on transient errors (rate limits, server errors).
func fetchWorkflows(client github.Client, repo github.Repository) tea.Cmd {
	return func() tea.Msg {
		var workflows []github.Workflow
		err := github.RetryWithBackoff(context.Background(), 3, func() error {
			var e error
			workflows, e = client.ListWorkflows(context.Background(), repo)
			return e
		})
		return WorkflowsLoadedMsg{
			Workflows: workflows,
			Err:       err,
		}
	}
}

// fetchRuns creates a command to fetch runs for a workflow.
// It captures the client, repo, and workflowID to avoid race conditions.
// Retries on transient errors (rate limits, server errors).
func fetchRuns(client github.Client, repo github.Repository, workflowID int64) tea.Cmd {
	return func() tea.Msg {
		opts := &github.ListRunsOpts{
			WorkflowID: workflowID,
		}
		var runs []github.Run
		err := github.RetryWithBackoff(context.Background(), 3, func() error {
			var e error
			runs, e = client.ListRuns(context.Background(), repo, opts)
			return e
		})
		return RunsLoadedMsg{
			Runs: runs,
			Err:  err,
		}
	}
}

// fetchJobs creates a command to fetch jobs for a run.
// It captures the client, repo, and runID to avoid race conditions.
// Retries on transient errors (rate limits, server errors).
func fetchJobs(client github.Client, repo github.Repository, runID int64) tea.Cmd {
	return func() tea.Msg {
		var jobs []github.Job
		err := github.RetryWithBackoff(context.Background(), 3, func() error {
			var e error
			jobs, e = client.ListJobs(context.Background(), repo, runID)
			return e
		})
		return JobsLoadedMsg{
			Jobs: jobs,
			Err:  err,
		}
	}
}

// fetchLogs creates a command to fetch logs for a job.
// It captures the client, repo, and jobID to avoid race conditions.
// Logs are sanitized to remove potential secrets before display.
// Retries on transient errors (rate limits, server errors).
func fetchLogs(client github.Client, repo github.Repository, jobID int64) tea.Cmd {
	return func() tea.Msg {
		var logs string
		err := github.RetryWithBackoff(context.Background(), 3, func() error {
			var e error
			logs, e = client.GetJobLogs(context.Background(), repo, jobID)
			return e
		})
		if err == nil {
			logs = github.SanitizeLogs(logs)
		}
		return LogsLoadedMsg{
			JobID: jobID,
			Logs:  logs,
			Err:   err,
		}
	}
}

// cancelRun creates a command to cancel a run.
// It captures the client, repo, and runID to avoid race conditions.
func cancelRun(client github.Client, repo github.Repository, runID int64) tea.Cmd {
	return func() tea.Msg {
		err := client.CancelRun(context.Background(), repo, runID)
		return RunCancelledMsg{
			RunID: runID,
			Err:   err,
		}
	}
}

// rerunWorkflow creates a command to rerun a workflow.
// It captures the client, repo, and runID to avoid race conditions.
func rerunWorkflow(client github.Client, repo github.Repository, runID int64) tea.Cmd {
	return func() tea.Msg {
		err := client.RerunWorkflow(context.Background(), repo, runID)
		return RunRerunMsg{
			RunID: runID,
			Err:   err,
		}
	}
}

// rerunFailedJobs creates a command to rerun only failed jobs.
// It captures the client, repo, and runID to avoid race conditions.
func rerunFailedJobs(client github.Client, repo github.Repository, runID int64) tea.Cmd {
	return func() tea.Msg {
		err := client.RerunFailedJobs(context.Background(), repo, runID)
		return RerunFailedJobsMsg{
			RunID: runID,
			Err:   err,
		}
	}
}

// triggerWorkflow creates a command to trigger a workflow dispatch.
// It captures the client, repo, workflowFile, ref, and inputs to avoid race conditions.
func triggerWorkflow(client github.Client, repo github.Repository, workflowFile string, ref string, inputs map[string]string) tea.Cmd {
	// Convert map[string]string to map[string]interface{} for the API
	inputsInterface := make(map[string]interface{}, len(inputs))
	for k, v := range inputs {
		inputsInterface[k] = v
	}

	return func() tea.Msg {
		err := client.TriggerWorkflow(context.Background(), repo, workflowFile, ref, inputsInterface)
		return WorkflowTriggeredMsg{
			Workflow: workflowFile,
			Err:      err,
		}
	}
}

// flashMessage creates a flash message that clears after duration.
// It returns a batch of commands: the flash message and a delayed clear.
func flashMessage(msg string, duration time.Duration) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return FlashMsg{
				Message:  msg,
				Duration: duration,
			}
		},
		tea.Tick(duration, func(t time.Time) tea.Msg {
			return FlashClearMsg{}
		}),
	)
}

// tick creates a tick command for polling.
// It waits for the specified interval before sending a TickMsg.
func tick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}
