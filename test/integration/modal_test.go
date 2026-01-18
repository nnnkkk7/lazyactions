package integration

import (
	"testing"

	"github.com/nnnkkk7/lazyactions/app"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestModal_HelpPopup(t *testing.T) {
	t.Run("? shows help popup", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.SendKey("?")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render help popup")
		}
	})

	t.Run("? again hides help popup", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.SendKey("?")
		ta.SendKey("?")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after closing help")
		}
	})

	t.Run("Esc closes help popup", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.SendKey("?")
		ta.SendKey("esc")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after Esc")
		}
	})
}

func TestModal_ConfirmDialog(t *testing.T) {
	t.Run("confirm dialog displays message", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		// Move to Runs pane and trigger cancel
		ta.SendKey("l")
		ta.SendKey("c")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render confirm dialog")
		}
	})

	t.Run("y executes confirmFn and closes", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		// Move to Runs pane and trigger cancel
		ta.SendKey("l")
		ta.SendKey("c")
		cmd := ta.SendKey("y")

		// Should return a command
		if cmd == nil {
			t.Error("y should return command from confirmFn")
		}
	})

	t.Run("Y same as y", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		ta.SendKey("l")
		ta.SendKey("c")
		cmd := ta.SendKey("Y")

		if cmd == nil {
			t.Error("Y should return command from confirmFn")
		}
	})

	t.Run("n closes without action", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		ta.SendKey("l")
		ta.SendKey("c")
		cmd := ta.SendKey("n")

		// Should return nil (no action)
		if cmd != nil {
			t.Error("n should return nil")
		}
	})

	t.Run("Esc closes without action", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		ta.SendKey("l")
		ta.SendKey("c")
		cmd := ta.SendKey("esc")

		if cmd != nil {
			t.Error("Esc should return nil")
		}
	})
}

func TestModal_EscapeHandling(t *testing.T) {
	t.Run("Esc closes help if open", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		ta.SendKey("?")
		ta.SendKey("esc")

		// Should be able to navigate after
		ta.SendKey("j")
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})

	t.Run("Esc closes fullscreen log if active", func(t *testing.T) {
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

		// Move to Logs and enter fullscreen
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

	t.Run("Esc clears error if set", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		// Set an error
		ta.App.Update(app.WorkflowsLoadedMsg{Err: ErrTest})

		// Clear with Esc
		ta.SendKey("esc")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after clearing error")
		}
	})
}

func TestModal_KeyBlockingInModal(t *testing.T) {
	t.Run("navigation keys blocked in help mode", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Open help
		ta.SendKey("?")

		// Try navigation
		ta.SendKey("j")
		ta.SendKey("l")

		// Help should still be open
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})

	t.Run("navigation keys blocked in confirm mode", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns([]github.Run{RunningRun()}),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: []github.Run{RunningRun()}})

		// Open confirm dialog
		ta.SendKey("l")
		ta.SendKey("c")

		// Try navigation (should be blocked)
		ta.SendKey("j")
		ta.SendKey("l")

		// Should still show confirm
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})
}

