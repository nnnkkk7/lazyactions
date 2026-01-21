package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestApp_HandleKeyPress_Quit(t *testing.T) {
	app := New()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	cmd := app.handleKeyPress(msg)

	// Quit should return tea.Quit command
	if cmd == nil {
		t.Error("Quit key should return a command")
	}
}

func TestApp_HandleKeyPress_Help(t *testing.T) {
	app := New()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	app.handleKeyPress(msg)

	if !app.showHelp {
		t.Error("? key should toggle showHelp to true")
	}

	app.handleKeyPress(msg)
	if app.showHelp {
		t.Error("? key should toggle showHelp back to false")
	}
}

func TestApp_HandleKeyPress_Escape(t *testing.T) {
	app := New()
	app.showHelp = true

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	app.handleKeyPress(msg)

	if app.showHelp {
		t.Error("Escape should close help")
	}
}

func TestApp_HandleKeyPress_Navigation(t *testing.T) {
	app := New()
	app.workflows.SetItems([]github.Workflow{
		{ID: 1, Name: "CI"},
		{ID: 2, Name: "Deploy"},
	})

	// Navigate down with arrow key
	msg := tea.KeyMsg{Type: tea.KeyDown}
	app.handleKeyPress(msg)

	if app.workflows.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", app.workflows.SelectedIndex())
	}

	// Navigate up with arrow key
	msg = tea.KeyMsg{Type: tea.KeyUp}
	app.handleKeyPress(msg)

	if app.workflows.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0", app.workflows.SelectedIndex())
	}
}

func TestApp_HandleKeyPress_PaneNavigation(t *testing.T) {
	app := New()

	// Move right to RunsPane
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	app.handleKeyPress(msg)

	if app.focusedPane != RunsPane {
		t.Errorf("focusedPane = %v, want RunsPane", app.focusedPane)
	}

	// Move right to JobsPane
	app.handleKeyPress(msg)

	if app.focusedPane != JobsPane {
		t.Errorf("focusedPane = %v, want JobsPane", app.focusedPane)
	}

	// Move left back to RunsPane
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	app.handleKeyPress(msg)

	if app.focusedPane != RunsPane {
		t.Errorf("focusedPane = %v, want RunsPane", app.focusedPane)
	}
}

func TestApp_HandleKeyPress_PanelNavigation_JK(t *testing.T) {
	app := New()

	// Initial state: WorkflowsPane
	if app.focusedPane != WorkflowsPane {
		t.Errorf("Initial focusedPane = %v, want WorkflowsPane", app.focusedPane)
	}

	// j moves to next pane (RunsPane)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	app.handleKeyPress(msg)

	if app.focusedPane != RunsPane {
		t.Errorf("After j: focusedPane = %v, want RunsPane", app.focusedPane)
	}

	// j moves to next pane (JobsPane)
	app.handleKeyPress(msg)

	if app.focusedPane != JobsPane {
		t.Errorf("After second j: focusedPane = %v, want JobsPane", app.focusedPane)
	}

	// j at JobsPane stays at JobsPane
	app.handleKeyPress(msg)

	if app.focusedPane != JobsPane {
		t.Errorf("After third j: focusedPane = %v, want JobsPane", app.focusedPane)
	}

	// k moves to previous pane (RunsPane)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	app.handleKeyPress(msg)

	if app.focusedPane != RunsPane {
		t.Errorf("After k: focusedPane = %v, want RunsPane", app.focusedPane)
	}

	// k moves to previous pane (WorkflowsPane)
	app.handleKeyPress(msg)

	if app.focusedPane != WorkflowsPane {
		t.Errorf("After second k: focusedPane = %v, want WorkflowsPane", app.focusedPane)
	}

	// k at WorkflowsPane stays at WorkflowsPane
	app.handleKeyPress(msg)

	if app.focusedPane != WorkflowsPane {
		t.Errorf("After third k: focusedPane = %v, want WorkflowsPane", app.focusedPane)
	}
}

func TestApp_HandleKeyPress_Filter(t *testing.T) {
	app := New()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	app.handleKeyPress(msg)

	if !app.filtering {
		t.Error("/ key should enable filtering mode")
	}
}

func TestApp_HandleKeyPress_FullLog(t *testing.T) {
	app := New()
	app.focusedPane = JobsPane

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}
	app.handleKeyPress(msg)

	if !app.fullscreenLog {
		t.Error("L key in JobsPane should enable fullscreenLog")
	}
}

func TestApp_HandleKeyPress_DetailTabSwitch(t *testing.T) {
	app := New()

	// Default should be LogsTab
	if app.detailTab != LogsTab {
		t.Errorf("Initial detailTab = %v, want LogsTab", app.detailTab)
	}

	// Press 1 to switch to InfoTab
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	app.handleKeyPress(msg)

	if app.detailTab != InfoTab {
		t.Errorf("After pressing 1: detailTab = %v, want InfoTab", app.detailTab)
	}

	// Press 2 to switch back to LogsTab
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	app.handleKeyPress(msg)

	if app.detailTab != LogsTab {
		t.Errorf("After pressing 2: detailTab = %v, want LogsTab", app.detailTab)
	}
}

func TestApp_HandleFilterInput_Escape(t *testing.T) {
	app := New()
	app.filtering = true

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	app.handleFilterInput(msg)

	if app.filtering {
		t.Error("Escape should disable filtering mode")
	}
}

func TestApp_HandleFilterInput_Enter(t *testing.T) {
	app := New()
	app.filtering = true
	app.filterInput.SetValue("test")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	app.handleFilterInput(msg)

	if app.filtering {
		t.Error("Enter should disable filtering mode")
	}
}

func TestApp_HandleConfirmInput_Yes(t *testing.T) {
	app := New()
	app.showConfirm = true
	called := false
	app.confirmFn = func() tea.Cmd {
		called = true
		return nil
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	app.handleConfirmInput(msg)

	if app.showConfirm {
		t.Error("y should close confirm dialog")
	}
	if !called {
		t.Error("y should call confirmFn")
	}
}

func TestApp_HandleConfirmInput_No(t *testing.T) {
	app := New()
	app.showConfirm = true
	app.confirmFn = func() tea.Cmd { return nil }

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	app.handleConfirmInput(msg)

	if app.showConfirm {
		t.Error("n should close confirm dialog")
	}
	if app.confirmFn != nil {
		t.Error("n should clear confirmFn")
	}
}

func TestApp_ApplyFilter(t *testing.T) {
	app := New()
	app.workflows.SetItems([]github.Workflow{
		{ID: 1, Name: "CI"},
		{ID: 2, Name: "Deploy"},
	})

	app.focusedPane = WorkflowsPane
	app.applyFilter("ci")

	if app.workflows.Len() != 1 {
		t.Errorf("workflows.Len() = %d, want 1 after filter", app.workflows.Len())
	}
}

func TestApp_FocusPrevPane(t *testing.T) {
	tests := []struct {
		start Pane
		want  Pane
	}{
		{WorkflowsPane, WorkflowsPane}, // Can't go before first
		{RunsPane, WorkflowsPane},
		{JobsPane, RunsPane},
	}

	for _, tt := range tests {
		app := New()
		app.focusedPane = tt.start
		app.focusPrevPane()

		if app.focusedPane != tt.want {
			t.Errorf("focusPrevPane from %v = %v, want %v", tt.start, app.focusedPane, tt.want)
		}
	}
}

func TestApp_FocusNextPane(t *testing.T) {
	tests := []struct {
		start Pane
		want  Pane
	}{
		{WorkflowsPane, RunsPane},
		{RunsPane, JobsPane},
		{JobsPane, JobsPane}, // Can't go past last
	}

	for _, tt := range tests {
		app := New()
		app.focusedPane = tt.start
		app.focusNextPane()

		if app.focusedPane != tt.want {
			t.Errorf("focusNextPane from %v = %v, want %v", tt.start, app.focusedPane, tt.want)
		}
	}
}

func TestApp_NavigateDownInRunsPane(t *testing.T) {
	mock := newMockClient(nil)
	app := New(WithClient(mock))
	app.focusedPane = RunsPane
	app.runs.SetItems([]github.Run{
		{ID: 1, Name: "Run 1"},
		{ID: 2, Name: "Run 2"},
	})

	app.navigateDown()

	if app.runs.SelectedIndex() != 1 {
		t.Errorf("runs.SelectedIndex() = %d, want 1", app.runs.SelectedIndex())
	}
}

func TestApp_NavigateUpInJobsPane(t *testing.T) {
	mock := newMockClient(nil)
	app := New(WithClient(mock))
	app.focusedPane = JobsPane
	app.jobs.SetItems([]github.Job{
		{ID: 1, Name: "build"},
		{ID: 2, Name: "test"},
	})
	app.jobs.SelectNext()

	app.navigateUp()

	if app.jobs.SelectedIndex() != 0 {
		t.Errorf("jobs.SelectedIndex() = %d, want 0", app.jobs.SelectedIndex())
	}
}

func TestApp_ApplyFilterOnRunsPane(t *testing.T) {
	app := New()
	app.focusedPane = RunsPane
	app.runs.SetItems([]github.Run{
		{ID: 1, Branch: "main"},
		{ID: 2, Branch: "feature"},
	})

	app.applyFilter("main")

	if app.runs.Len() != 1 {
		t.Errorf("runs.Len() = %d, want 1 after filter", app.runs.Len())
	}
}

func TestApp_ApplyFilterOnJobsPane(t *testing.T) {
	app := New()
	app.focusedPane = JobsPane
	app.jobs.SetItems([]github.Job{
		{ID: 1, Name: "build"},
		{ID: 2, Name: "test"},
	})

	app.applyFilter("build")

	if app.jobs.Len() != 1 {
		t.Errorf("jobs.Len() = %d, want 1 after filter", app.jobs.Len())
	}
}
