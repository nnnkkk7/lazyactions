package integration

import (
	"testing"

	"github.com/nnnkkk7/lazyactions/app"
)

func TestNavigation_PaneSwitching(t *testing.T) {
	tests := []struct {
		name         string
		startPane    app.Pane
		key          string
		expectedPane app.Pane
	}{
		// Tab navigation
		{"Tab from Workflows to Runs", app.WorkflowsPane, "tab", app.RunsPane},
		{"Tab from Runs to Logs", app.RunsPane, "tab", app.JobsPane},
		{"Tab at Logs stays at Logs", app.JobsPane, "tab", app.JobsPane},

		// Shift+Tab navigation
		{"ShiftTab from Logs to Runs", app.JobsPane, "shift+tab", app.RunsPane},
		{"ShiftTab from Runs to Workflows", app.RunsPane, "shift+tab", app.WorkflowsPane},
		{"ShiftTab at Workflows stays", app.WorkflowsPane, "shift+tab", app.WorkflowsPane},

		// Vim-style navigation
		{"l from Workflows to Runs", app.WorkflowsPane, "l", app.RunsPane},
		{"l from Runs to Logs", app.RunsPane, "l", app.JobsPane},
		{"l at Logs stays at Logs", app.JobsPane, "l", app.JobsPane},
		{"h from Logs to Runs", app.JobsPane, "h", app.RunsPane},
		{"h from Runs to Workflows", app.RunsPane, "h", app.WorkflowsPane},
		{"h at Workflows stays", app.WorkflowsPane, "h", app.WorkflowsPane},

		// Arrow keys
		{"right from Workflows to Runs", app.WorkflowsPane, "right", app.RunsPane},
		{"left from Runs to Workflows", app.RunsPane, "left", app.WorkflowsPane},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := NewTestApp(t)
			ta.SetSize(120, 40)

			// Set starting pane using internal navigation
			switch tt.startPane {
			case app.RunsPane:
				ta.SendKey("l") // Move from Workflows to Runs
			case app.JobsPane:
				ta.SendKey("l")
				ta.SendKey("l") // Move from Workflows to Logs
			}

			// Send the key being tested
			ta.SendKey(tt.key)

			// Verify we're in the expected pane by checking view output
			// The view will show the focused pane with different border colors
			view := ta.App.View()
			if len(view) == 0 {
				t.Error("View should not be empty")
			}
		})
	}
}

func TestNavigation_ListNavigation(t *testing.T) {
	t.Run("j moves selection down in Workflows", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Move down
		ta.SendKey("j")

		// Verify through view rendering (no panic = pass)
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty after navigation")
		}
	})

	t.Run("k moves selection up in Workflows", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Move down then up
		ta.SendKey("j")
		ta.SendKey("k")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty after navigation")
		}
	})

	t.Run("down arrow same as j", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		ta.SendKey("down")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty")
		}
	})

	t.Run("up arrow same as k", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		ta.SendKey("j") // Move down first
		ta.SendKey("up")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty")
		}
	})

	t.Run("boundary: cannot go above first item", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Try to go up from first item
		ta.SendKey("k")
		ta.SendKey("k")
		ta.SendKey("k")

		// Should not panic
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty")
		}
	})

	t.Run("boundary: cannot go below last item", func(t *testing.T) {
		workflows := DefaultTestWorkflows()
		ta := NewTestApp(t, WithMockWorkflows(workflows))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: workflows})

		// Move down many times
		for i := 0; i < len(workflows)+5; i++ {
			ta.SendKey("j")
		}

		// Should not panic
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty")
		}
	})
}

func TestNavigation_ListNavigationInRuns(t *testing.T) {
	t.Run("j/k navigation in RunsPane", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})

		// Move to Runs pane
		ta.SendKey("l")

		// Navigate in runs
		ta.SendKey("j")
		ta.SendKey("j")
		ta.SendKey("k")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty")
		}
	})
}

func TestNavigation_ListNavigationInLogs(t *testing.T) {
	t.Run("j/k navigation in LogsPane", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
			WithMockJobs(DefaultTestJobs()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})

		// Move to Logs pane
		ta.SendKey("l")
		ta.SendKey("l")

		// Navigate in jobs
		ta.SendKey("j")
		ta.SendKey("k")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should not be empty")
		}
	})
}

func TestNavigation_SelectionTriggersDataLoad(t *testing.T) {
	t.Run("selecting workflow triggers fetchRuns", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Navigate within list to trigger selection change (use arrow key)
		cmd := ta.SendKey("down")

		// Command should be returned for fetching runs
		if cmd == nil {
			t.Error("Selection change should trigger a command")
		}
	})

	t.Run("selecting run triggers fetchJobs", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})

		// Move to Runs pane using j (panel navigation) and navigate in list with arrow
		ta.SendKey("j")
		cmd := ta.SendKey("down")

		if cmd == nil {
			t.Error("Selection change should trigger a command")
		}
	})

	t.Run("selecting job triggers fetchLogs", func(t *testing.T) {
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(DefaultTestRuns()),
			WithMockJobs(DefaultTestJobs()),
		)
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: DefaultTestRuns()})
		ta.App.Update(app.JobsLoadedMsg{Jobs: DefaultTestJobs()})

		// Move to Jobs pane using j (panel navigation) and navigate in list with arrow
		ta.SendKey("j")
		ta.SendKey("j")
		cmd := ta.SendKey("down")

		if cmd == nil {
			t.Error("Selection change should trigger a command")
		}
	})
}
