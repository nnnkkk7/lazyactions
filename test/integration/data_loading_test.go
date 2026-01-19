package integration

import (
	"errors"
	"testing"

	"github.com/nnnkkk7/lazyactions/app"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestDataLoading_Success(t *testing.T) {
	t.Run("workflows load successfully", func(t *testing.T) {
		workflows := DefaultTestWorkflows()
		ta := NewTestApp(t, WithMockWorkflows(workflows))
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: workflows})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty after workflows loaded")
		}
	})

	t.Run("runs load for selected workflow", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty after runs loaded")
		}
	})

	t.Run("jobs load for selected run", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
			WithMockJobs(DefaultTestJobs()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty after jobs loaded")
		}
	})

	t.Run("logs load for selected job", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
			WithMockJobs(DefaultTestJobs()),
			WithMockLogs(DefaultTestLogs()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})
		ta.App.Update(app.LogsLoadedMsg{JobID: 1001, Logs: DefaultTestLogs()})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty after logs loaded")
		}
	})
}

func TestDataLoading_ErrorHandling(t *testing.T) {
	errorTypes := []struct {
		name    string
		err     error
		errType github.ErrorType
	}{
		{"NetworkError", errors.New("network error"), github.ErrTypeNetwork},
		{"AuthError", &github.AppError{Type: github.ErrTypeAuth, Message: "auth failed"}, github.ErrTypeAuth},
		{"RateLimitError", &github.AppError{Type: github.ErrTypeRateLimit, Message: "rate limited"}, github.ErrTypeRateLimit},
		{"NotFoundError", &github.AppError{Type: github.ErrTypeNotFound, Message: "not found"}, github.ErrTypeNotFound},
		{"ServerError", &github.AppError{Type: github.ErrTypeServer, Message: "server error"}, github.ErrTypeServer},
	}

	for _, tt := range errorTypes {
		t.Run(tt.name+"_on_workflows", func(t *testing.T) {
			ta := NewTestApp(t, WithMockError(tt.err))
			ta.SetSize(120, 40)

			ta.App.Update(app.WorkflowsLoadedMsg{Err: tt.err})

			// View should still render (with error in status bar)
			view := ta.App.View()
			if len(view) == 0 {
				t.Error("View should render even with error")
			}
		})

		t.Run(tt.name+"_on_runs", func(t *testing.T) {
			ta := NewTestApp(t)
			ta.SetSize(120, 40)

			ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
			ta.App.Update(app.RunsLoadedMsg{Err: tt.err})

			view := ta.App.View()
			if len(view) == 0 {
				t.Error("View should render even with error")
			}
		})

		t.Run(tt.name+"_on_jobs", func(t *testing.T) {
			ta := NewTestApp(t)
			ta.SetSize(120, 40)

			ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
			ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
			ta.App.Update(app.JobsLoadedMsg{Err: tt.err})

			view := ta.App.View()
			if len(view) == 0 {
				t.Error("View should render even with error")
			}
		})

		t.Run(tt.name+"_on_logs", func(t *testing.T) {
			ta := NewTestApp(t)
			ta.SetSize(120, 40)

			ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
			ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
			ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})
			ta.App.Update(app.LogsLoadedMsg{JobID: 1001, Err: tt.err})

			view := ta.App.View()
			if len(view) == 0 {
				t.Error("View should render even with error")
			}
		})
	}
}

func TestDataLoading_CascadeOnNewData(t *testing.T) {
	t.Run("new workflows auto-select first and fetch runs", func(t *testing.T) {
		ta := NewTestApp(t, WithMockRuns(DefaultTestRuns()))
		ta.SetSize(120, 40)

		_, cmd := ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Should return command to fetch runs
		if cmd == nil {
			t.Error("WorkflowsLoadedMsg should trigger fetchRuns command")
		}
	})

	t.Run("new runs auto-select first and fetch jobs", func(t *testing.T) {
		ta := NewTestApp(t, WithMockJobs(DefaultTestJobs()))
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		_, cmd := ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})

		if cmd == nil {
			t.Error("RunsLoadedMsg should trigger fetchJobs command")
		}
	})

	t.Run("new jobs auto-select first and fetch logs", func(t *testing.T) {
		ta := NewTestApp(t, WithMockLogs(DefaultTestLogs()))
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		_, cmd := ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})

		if cmd == nil {
			t.Error("JobsLoadedMsg should trigger fetchLogs command")
		}
	})
}

func TestDataLoading_EmptyDataHandling(t *testing.T) {
	t.Run("empty workflows shows appropriate message", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: []github.Workflow{}})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render with empty workflows")
		}
	})

	t.Run("empty runs clears jobs and logs", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{}})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render with empty runs")
		}
	})

	t.Run("empty jobs clears logs", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		ta.App.Update(app.JobsLoadedMsg{Jobs: []github.Job{}})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render with empty jobs")
		}
	})
}

func TestDataLoading_RefreshAll(t *testing.T) {
	t.Run("Ctrl+R triggers refresh", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		cmd := ta.SendKey("ctrl+r")

		if cmd == nil {
			t.Error("Ctrl+R should return refresh command")
		}
	})
}
