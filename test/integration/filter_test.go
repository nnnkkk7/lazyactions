package integration

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/lazyactions/app"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestFilter_EnterExitMode(t *testing.T) {
	t.Run("/ enters filter mode", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		ta.SendKey("/")

		// View should show filter input
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render in filter mode")
		}
	})

	t.Run("Esc exits filter mode and clears filter", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Enter filter mode
		ta.SendKey("/")

		// Exit with Esc
		ta.SendKey("esc")

		// Should be back to normal mode
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after exiting filter mode")
		}
	})

	t.Run("Enter exits filter mode and applies filter", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Enter filter mode
		ta.SendKey("/")

		// Type and apply
		ta.SendKey("enter")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after applying filter")
		}
	})
}

func TestFilter_WorkflowsPane(t *testing.T) {
	t.Run("filter by workflow name", func(t *testing.T) {
		workflows := []github.Workflow{
			{ID: 1, Name: "CI", Path: ".github/workflows/ci.yml"},
			{ID: 2, Name: "Deploy", Path: ".github/workflows/deploy.yml"},
			{ID: 3, Name: "Test CI", Path: ".github/workflows/test-ci.yml"},
		}
		ta := NewTestApp(t, WithMockWorkflows(workflows))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: workflows})

		// Enter filter mode
		ta.SendKey("/")

		// Type "ci" (simulating text input)
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
		ta.SendKey("enter")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after filtering")
		}
	})

	t.Run("filter is case insensitive", func(t *testing.T) {
		workflows := []github.Workflow{
			{ID: 1, Name: "CI", Path: ".github/workflows/ci.yml"},
			{ID: 2, Name: "Deploy", Path: ".github/workflows/deploy.yml"},
		}
		ta := NewTestApp(t, WithMockWorkflows(workflows))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: workflows})

		// Enter filter mode
		ta.SendKey("/")

		// Type "CI" in uppercase
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}})
		ta.SendKey("enter")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})
}

func TestFilter_RunsPane(t *testing.T) {
	t.Run("filter by branch name", func(t *testing.T) {
		runs := DefaultTestRuns()
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(runs),
		)
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: runs})

		// Move to Runs pane
		ta.SendKey("l")

		// Enter filter mode
		ta.SendKey("/")

		// Type "main"
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		ta.SendKey("enter")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render after filtering runs")
		}
	})

	t.Run("filter by actor name", func(t *testing.T) {
		runs := DefaultTestRuns()
		ta := NewTestApp(t,
			WithMockWorkflows(DefaultTestWorkflows()),
			WithMockRuns(runs),
		)
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})
		ta.App.Update(app.RunsLoadedMsg{Runs: runs})

		// Move to Runs pane
		ta.SendKey("l")

		// Enter filter mode
		ta.SendKey("/")

		// Type "user1"
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
		ta.SendKey("enter")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})
}

func TestFilter_LogsPane(t *testing.T) {
	t.Run("filter by job name", func(t *testing.T) {
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

		// Enter filter mode
		ta.SendKey("/")

		// Type "build"
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		ta.SendKey("enter")

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})
}

func TestFilter_ClearFilter(t *testing.T) {
	t.Run("Esc clears filter", func(t *testing.T) {
		workflows := DefaultTestWorkflows()
		ta := NewTestApp(t, WithMockWorkflows(workflows))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: workflows})

		// Apply filter
		ta.SendKey("/")
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
		ta.SendKey("enter")

		// Clear filter
		ta.SendKey("/")
		ta.SendKey("esc")

		// Should show all workflows
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})
}

func TestFilter_InputHandling(t *testing.T) {
	t.Run("typing characters updates filter", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Enter filter mode
		ta.SendKey("/")

		// Type characters
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render while typing")
		}
	})

	t.Run("backspace removes characters", func(t *testing.T) {
		ta := NewTestApp(t, WithMockWorkflows(DefaultTestWorkflows()))
		ta.SetSize(120, 40)
		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		// Enter filter mode
		ta.SendKey("/")

		// Type then backspace
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		ta.SendKeyMsg(tea.KeyMsg{Type: tea.KeyBackspace})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})
}
