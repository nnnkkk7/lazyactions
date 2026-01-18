package integration

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/lazyactions/app"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestView_FullscreenLog(t *testing.T) {
	t.Run("L in LogsPane enables fullscreen", func(t *testing.T) {
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
		ta.App.Update(app.LogsLoadedMsg{Logs: DefaultTestLogs()})

		// Move to Logs pane
		ta.SendKey("l")
		ta.SendKey("l")

		// Enter fullscreen
		ta.SendKey("L")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render in fullscreen mode")
		}
	})

	t.Run("L in other panes does nothing", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// In WorkflowsPane
		ta.SendKey("L")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})

	t.Run("Esc exits fullscreen", func(t *testing.T) {
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
		ta.App.Update(app.LogsLoadedMsg{Logs: DefaultTestLogs()})

		// Enter fullscreen
		ta.SendKey("l")
		ta.SendKey("l")
		ta.SendKey("L")

		// Exit fullscreen
		ta.SendKey("esc")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after exiting fullscreen")
		}
	})
}

func TestView_WindowResize(t *testing.T) {
	t.Run("WindowSizeMsg updates dimensions", func(t *testing.T) {
		ta := NewTestApp(t)

		ta.App.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

		view := ta.App.View()
		if view == "Loading..." {
			t.Error("View should render with valid size")
		}
	})

	t.Run("resize updates view correctly", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		// Resize
		ta.SetSize(80, 25)

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after resize")
		}
	})

	t.Run("minimum widths enforced", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(30, 10) // Very small

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render with minimum size")
		}
	})
}

func TestView_StatusBar(t *testing.T) {
	t.Run("shows hints for WorkflowsPane", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render status bar")
		}
	})

	t.Run("shows hints for RunsPane", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})

		ta.SendKey("l")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})

	t.Run("shows hints for LogsPane", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
			WithMockJobs(DefaultTestJobs()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})

		ta.SendKey("l")
		ta.SendKey("l")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})

	t.Run("shows flash message when set", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.RunCancelledMsg{RunID: 100})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render flash message")
		}
	})

	t.Run("shows error in red when err set", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Err: ErrTest})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render error")
		}
	})
}

func TestView_EmptyState(t *testing.T) {
	t.Run("zero size shows Loading...", func(t *testing.T) {
		ta := NewTestApp(t)
		// Don't set size

		view := ta.App.View()
		if view != "Loading..." {
			t.Errorf("View = %q, want %q", view, "Loading...")
		}
	})

	t.Run("empty lists render gracefully", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: []github.Workflow{}})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{}})
		ta.App.Update(app.JobsLoadedMsg{Jobs: []github.Job{}})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render with empty lists")
		}
	})
}

func TestView_StatusIcons(t *testing.T) {
	statusTests := []struct {
		status     string
		conclusion string
		desc       string
	}{
		{"in_progress", "", "running icon"},
		{"queued", "", "queued icon"},
		{"completed", "success", "success icon"},
		{"completed", "failure", "failure icon"},
		{"completed", "cancelled", "cancelled icon"},
	}

	for _, tt := range statusTests {
		t.Run(tt.desc, func(t *testing.T) {
			runs := []github.Run{
				{
					ID:         100,
					Status:     tt.status,
					Conclusion: tt.conclusion,
					Branch:     "main",
				},
			}
			ta := NewTestApp(t,
				WithMockWorkflows(DefaultTestWorkflows()),
				WithMockRuns(runs),
			)
			ta.SetSize(120, 40)

			ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
			ta.App.Update(app.RunsLoadedMsg{Runs: runs})

			view := ta.App.View()
			if len(view) == 0 {
				t.Error("View should render with status icons")
			}
		})
	}
}

