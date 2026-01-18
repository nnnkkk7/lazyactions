package integration

import (
	"testing"

	"github.com/nnnkkk7/lazyactions/app"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestIntegration_FullUserFlow(t *testing.T) {
	t.Run("browse workflows, view run, cancel it", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
			WithMockJobs(DefaultTestJobs()),
			WithMockLogs(DefaultTestLogs()),
		)
		ta.SetSize(120, 40)

		// 1. App starts and loads workflows
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// 2. Navigate to RunsPane
		ta.SendKey("l")
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		// 3. Select a running run (already selected)
		// 4. Press 'c' to cancel
		ta.SendKey("c")

		// 5. Confirm with 'y'
		cmd := ta.SendKey("y")

		if cmd == nil {
			t.Error("Cancel confirmation should return command")
		}

		// 6. Simulate cancel success
		ta.App.Update(app.RunCancelledMsg{RunID: 200})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after cancel flow")
		}
	})
}

func TestIntegration_FilterAndAction(t *testing.T) {
	t.Run("filter runs then rerun one", func(t *testing.T) {
		runs := []github.Run{
			{ID: 100, Branch: "main", Status: "completed", Conclusion: "success"},
			{ID: 101, Branch: "feature", Status: "completed", Conclusion: "failure"},
		}
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(runs),
		)
		ta.SetSize(120, 40)

		// Load data
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: runs})

		// 1. Navigate to RunsPane
		ta.SendKey("l")

		// 2. Enter filter mode
		ta.SendKey("/")

		// 3. Type filter text "main"
		for _, r := range "main" {
			ta.SendKeyMsg(keyMsgFromString(string(r)))
		}
		ta.SendKey("enter")

		// 4. Select filtered run (first one)
		// 5. Press 'r' to rerun
		cmd := ta.SendKey("r")

		if cmd == nil {
			t.Error("Rerun should return command")
		}

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after filter and action")
		}
	})
}

func TestIntegration_ErrorRecovery(t *testing.T) {
	t.Run("API error then recovery", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)

		// 1. API returns error
		ta.App.Update(app.WorkflowsLoadedMsg{Err: &github.AppError{
			Type:    github.ErrTypeNetwork,
			Message: "network error",
		}})

		// 2. Error displayed in status bar (view should render)
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render with error")
		}

		// 3. User presses Esc to clear error
		ta.SendKey("esc")

		// 4. User refreshes with Ctrl+R
		cmd := ta.SendKey("ctrl+r")

		if cmd == nil {
			t.Error("Refresh should return command")
		}

		// 5. Data loads successfully
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		view = ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after recovery")
		}
	})
}

func TestIntegration_MultiPaneWorkflow(t *testing.T) {
	t.Run("navigate through all panes viewing details", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
			WithMockJobs(DefaultTestJobs()),
			WithMockLogs(DefaultTestLogs()),
		)
		ta.SetSize(120, 40)

		// 1. Load workflows and select workflow
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// 2. Navigate to runs, load and select run
		ta.SendKey("l")
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})

		// 3. Navigate to logs, load and view job logs
		ta.SendKey("l")
		ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})
		ta.App.Update(app.LogsLoadedMsg{Logs: DefaultTestLogs()})

		// 4. Enter fullscreen log mode
		ta.SendKey("L")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render in fullscreen")
		}

		// 5. Exit and navigate back
		ta.SendKey("esc")
		ta.SendKey("h")
		ta.SendKey("h")

		view = ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after navigating back")
		}
	})
}

func TestIntegration_RapidNavigation(t *testing.T) {
	t.Run("rapid key presses handled correctly", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
			WithMockJobs(DefaultTestJobs()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})

		// Rapid navigation
		keys := []string{
			"j", "j", "j", "k", "k",
			"l", "j", "j", "k",
			"l", "j", "k",
			"h", "h",
			"j", "l", "l", "h", "k",
		}

		for _, key := range keys {
			ta.SendKey(key)
		}

		// Should not panic and state should be consistent
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after rapid navigation")
		}
	})
}

func TestIntegration_ConcurrentOperations(t *testing.T) {
	t.Run("multiple data loads don't corrupt state", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
			WithMockJobs(DefaultTestJobs()),
			WithMockLogs(DefaultTestLogs()),
		)
		ta.SetSize(120, 40)

		// Simulate multiple loads happening
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})
		ta.App.Update(app.LogsLoadedMsg{Logs: DefaultTestLogs()})

		// Simulate another round of loads
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()[:1]})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()[:2]})

		// View should still render correctly
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})
}

func TestIntegration_HelpDuringOperation(t *testing.T) {
	t.Run("can open help during any state", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Open help
		ta.SendKey("?")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render help")
		}

		// Close help
		ta.SendKey("?")

		// Navigate
		ta.SendKey("l")

		// Open help again
		ta.SendKey("?")

		view = ta.App.View()
		if len(view) == 0 {
			t.Error("View should render help in different pane")
		}
	})
}

func TestIntegration_EmptyToFilledData(t *testing.T) {
	t.Run("handle transition from empty to filled data", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		// Start with empty data
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: []github.Workflow{}})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render with empty data")
		}

		// Then receive data
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		view = ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after receiving data")
		}
	})
}
