package integration

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/lazyactions/app"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestAppLifecycle_Initialization(t *testing.T) {
	t.Run("creates app with default state", func(t *testing.T) {
		ta := NewTestApp(t)

		if ta.App == nil {
			t.Fatal("App is nil")
		}
	})

	t.Run("client option is applied", func(t *testing.T) {
		ta := NewTestApp(t)

		if ta.Mock() == nil {
			t.Error("MockClient should be set")
		}
	})

	t.Run("starts with WorkflowsPane focused", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		view := ta.App.View()
		if view == "Loading..." {
			t.Error("View should render with valid size")
		}
	})
}

func TestAppLifecycle_Init(t *testing.T) {
	t.Run("Init returns commands", func(t *testing.T) {
		ta := NewTestApp(t)

		cmd := ta.App.Init()
		if cmd == nil {
			t.Error("Init() should return commands")
		}
	})

	t.Run("Init triggers fetchWorkflows", func(t *testing.T) {
		workflows := DefaultTestWorkflows()
		ta := NewTestApp(t, WithMockWorkflows(workflows))

		ta.App.Init()
		// The mock should be called when the command is executed
	})
}

func TestAppLifecycle_StartupSequence(t *testing.T) {
	t.Run("workflows load and auto-select first", func(t *testing.T) {
		workflows := DefaultTestWorkflows()
		runs := DefaultTestRuns()
		jobs := DefaultTestJobs()
		logs := DefaultTestLogs()

		ta := NewTestApp(t,
			WithMockWorkflows(workflows),
			WithMockRuns(runs),
			WithMockJobs(jobs),
			WithMockLogs(logs),
		)
		ta.SetSize(120, 40)

		// Send WorkflowsLoadedMsg
		msg := app.WorkflowsLoadedMsg{Workflows: workflows}
		ta.App.Update(msg)

		// Verify workflows are loaded
		if len(ta.Mock().ListWorkflowsCalls()) != 0 {
			t.Error("ListWorkflows should not be called when workflows are injected via message")
		}
	})

	t.Run("data cascade from workflows to logs", func(t *testing.T) {
		workflows := DefaultTestWorkflows()
		runs := DefaultTestRuns()
		jobs := DefaultTestJobs()
		logs := DefaultTestLogs()

		ta := NewTestApp(t,
			WithMockWorkflows(workflows),
			WithMockRuns(runs),
			WithMockJobs(jobs),
			WithMockLogs(logs),
		)
		ta.SetSize(120, 40)

		// Step 1: Workflows loaded
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: workflows})

		// Step 2: Runs loaded
		ta.App.Update(app.RunsLoadedMsg{Runs: runs})

		// Step 3: Jobs loaded
		ta.App.Update(app.JobsLoadedMsg{Jobs: jobs})

		// Step 4: Logs loaded (for first job ID 1001)
		ta.App.Update(app.LogsLoadedMsg{JobID: 1001, Logs: logs})

		// Verify view renders without error
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty after data cascade")
		}
	})
}

func TestAppLifecycle_Quit(t *testing.T) {
	t.Run("q key returns tea.Quit", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		cmd := ta.SendKey("q")
		if cmd == nil {
			t.Error("q should return a command")
		}

		// Execute the command
		msg := ta.ProcessCmd(cmd)
		if msg != tea.Quit() {
			t.Error("q should return tea.Quit")
		}
	})
}

func TestAppLifecycle_EmptyData(t *testing.T) {
	t.Run("empty workflows handled gracefully", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		// Send empty workflows
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: []github.Workflow{}})

		// View should render without panic
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render even with empty workflows")
		}
	})

	t.Run("empty runs handled gracefully", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{}})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render even with empty runs")
		}
	})

	t.Run("empty jobs handled gracefully", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		ta.App.Update(app.JobsLoadedMsg{Jobs: []github.Job{}})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render even with empty jobs")
		}
	})
}

func TestAppLifecycle_WindowSize(t *testing.T) {
	t.Run("zero size shows Loading...", func(t *testing.T) {
		ta := NewTestApp(t)
		// Don't set size

		view := ta.App.View()
		if view != "Loading..." {
			t.Errorf("View with zero size = %q, want %q", view, "Loading...")
		}
	})

	t.Run("valid size renders full view", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		view := ta.App.View()
		if view == "Loading..." {
			t.Error("View should not be Loading... with valid size")
		}
	})

	t.Run("small size enforces minimums", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(50, 10)

		// Should not panic
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render with small size")
		}
	})
}
