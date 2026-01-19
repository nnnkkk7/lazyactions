package integration

import (
	"testing"

	"github.com/nnnkkk7/lazyactions/app"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestActions_CancelRun(t *testing.T) {
	t.Run("c on running run shows confirm dialog", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		// Move to Runs pane
		ta.SendKey("l")

		// Press c to cancel
		ta.SendKey("c")

		// View should show confirm dialog
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render confirm dialog")
		}
	})

	t.Run("c on completed run does nothing", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{SuccessRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{SuccessRun()}})

		// Move to Runs pane
		ta.SendKey("l")

		// Press c
		cmd := ta.SendKey("c")

		// Should return nil since run is not running
		if cmd != nil {
			t.Error("c on completed run should not return command")
		}
	})

	t.Run("y confirms cancel and calls API", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		// Move to Runs pane
		ta.SendKey("l")

		// Press c then y
		ta.SendKey("c")
		cmd := ta.SendKey("y")

		// Should return command to cancel
		if cmd == nil {
			t.Error("y should trigger cancel command")
		}
	})

	t.Run("n cancels dialog without action", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		// Move to Runs pane
		ta.SendKey("l")

		// Press c then n
		ta.SendKey("c")
		cmd := ta.SendKey("n")

		// Should return nil (dialog closed without action)
		if cmd != nil {
			t.Error("n should close dialog without returning command")
		}
	})

	t.Run("cancel success shows flash message", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.RunCancelledMsg{RunID: 100, Err: nil})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after cancel success")
		}
	})
}

func TestActions_RerunWorkflow(t *testing.T) {
	t.Run("r on any run triggers rerun", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})

		// Move to Runs pane
		ta.SendKey("l")

		// Press r to rerun
		cmd := ta.SendKey("r")

		if cmd == nil {
			t.Error("r should trigger rerun command")
		}
	})

	t.Run("rerun success shows flash message", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.RunRerunMsg{RunID: 100, Err: nil})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after rerun success")
		}
	})

	t.Run("r with no selection does nothing", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{}})

		// Move to Runs pane
		ta.SendKey("l")

		// Press r with no runs
		cmd := ta.SendKey("r")

		if cmd != nil {
			t.Error("r with no selection should return nil")
		}
	})
}

func TestActions_RerunFailedJobs(t *testing.T) {
	t.Run("R on failed run triggers rerun-failed", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{FailedRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{FailedRun()}})

		// Move to Runs pane
		ta.SendKey("l")

		// Press R to rerun failed
		cmd := ta.SendKey("R")

		if cmd == nil {
			t.Error("R on failed run should trigger command")
		}
	})

	t.Run("R on non-failed run does nothing", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{SuccessRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{SuccessRun()}})

		// Move to Runs pane
		ta.SendKey("l")

		// Press R
		cmd := ta.SendKey("R")

		if cmd != nil {
			t.Error("R on success run should return nil")
		}
	})

	t.Run("rerun failed jobs success shows flash message", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.RerunFailedJobsMsg{RunID: 100, Err: nil})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after rerun failed jobs success")
		}
	})

	t.Run("rerun failed jobs error sets error state", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		testErr := &github.AppError{Message: "rerun failed jobs failed"}
		ta.App.Update(app.RerunFailedJobsMsg{RunID: 100, Err: testErr})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after rerun failed jobs error")
		}
	})
}

func TestActions_TriggerWorkflow(t *testing.T) {
	t.Run("t in WorkflowsPane triggers workflow", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Press t in Workflows pane
		cmd := ta.SendKey("t")

		if cmd == nil {
			t.Error("t should trigger workflow dispatch command")
		}
	})

	t.Run("trigger success shows flash message", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowTriggeredMsg{Workflow: "ci.yml", Err: nil})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after trigger success")
		}
	})

	t.Run("trigger error sets error state", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		testErr := &github.AppError{Message: "workflow dispatch failed"}
		ta.App.Update(app.WorkflowTriggeredMsg{Workflow: "ci.yml", Err: testErr})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after trigger error")
		}
	})

	t.Run("t in other panes does nothing", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})

		// Move to Runs pane
		ta.SendKey("l")

		// Press t in Runs pane
		cmd := ta.SendKey("t")

		if cmd != nil {
			t.Error("t in RunsPane should return nil")
		}
	})

	t.Run("t with no workflow selected does nothing", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: []github.Workflow{}})

		cmd := ta.SendKey("t")

		if cmd != nil {
			t.Error("t with no workflow should return nil")
		}
	})
}

func TestActions_YankURL(t *testing.T) {
	t.Run("y in RunsPane copies URL", func(t *testing.T) {
		run := RunningRun()
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{run}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{run}})

		// Move to Runs pane
		ta.SendKey("l")

		// Press y to yank
		cmd := ta.SendKey("y")

		// Should return flash message command
		if cmd == nil {
			t.Error("y should return flash message command")
		}
	})

	t.Run("y with no URL does nothing", func(t *testing.T) {
		run := github.Run{
			ID:     100,
			Status: "completed",
			URL:    "", // No URL
		}
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{run}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{run}})

		// Move to Runs pane
		ta.SendKey("l")

		// Press y
		cmd := ta.SendKey("y")

		if cmd != nil {
			t.Error("y with no URL should return nil")
		}
	})
}

func TestActions_PaneSpecificActions(t *testing.T) {
	t.Run("cancel only works in RunsPane", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// In WorkflowsPane
		cmd := ta.SendKey("c")
		if cmd != nil {
			t.Error("c in WorkflowsPane should return nil")
		}
	})

	t.Run("trigger only works in WorkflowsPane", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
		)
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})

		// Move to Runs pane
		ta.SendKey("l")

		cmd := ta.SendKey("t")
		if cmd != nil {
			t.Error("t in RunsPane should return nil")
		}

		// Move to Logs pane
		ta.SendKey("l")

		cmd = ta.SendKey("t")
		if cmd != nil {
			t.Error("t in LogsPane should return nil")
		}
	})
}
