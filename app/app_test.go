package app

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestNew_CreatesApp(t *testing.T) {
	app := New()

	if app == nil {
		t.Fatal("New() returned nil")
	}
	if app.workflows == nil {
		t.Error("workflows is nil")
	}
	if app.runs == nil {
		t.Error("runs is nil")
	}
	if app.jobs == nil {
		t.Error("jobs is nil")
	}
	if app.logView == nil {
		t.Error("logView is nil")
	}
	if app.focusedPane != WorkflowsPane {
		t.Errorf("focusedPane = %v, want WorkflowsPane", app.focusedPane)
	}
}

func TestNew_WithClient(t *testing.T) {
	mock := newMockClient(nil)
	app := New(WithClient(mock))

	if app.client != mock {
		t.Error("WithClient option not applied")
	}
	if app.adaptivePoller == nil {
		t.Error("adaptivePoller should be created when client is set")
	}
}

func TestNew_WithRepository(t *testing.T) {
	repo := github.Repository{Owner: "test", Name: "repo"}
	app := New(WithRepository(repo))

	if app.repo.Owner != "test" || app.repo.Name != "repo" {
		t.Errorf("repo = %v, want %v", app.repo, repo)
	}
}

func TestApp_Init(t *testing.T) {
	app := New()
	cmd := app.Init()

	if cmd == nil {
		t.Error("Init() returned nil cmd")
	}
}

func TestApp_View_ZeroSize(t *testing.T) {
	app := New()
	view := app.View()

	if view != "Loading..." {
		t.Errorf("View() with zero size = %q, want %q", view, "Loading...")
	}
}

func TestApp_View_WithSize(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40

	view := app.View()

	if view == "Loading..." {
		t.Error("View() should not return Loading... with valid size")
	}
	if len(view) == 0 {
		t.Error("View() returned empty string")
	}
}

func TestApp_View_HelpMode(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.showHelp = true

	view := app.View()

	if len(view) == 0 {
		t.Error("View() in help mode returned empty string")
	}
}

func TestApp_View_ConfirmMode(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.showConfirm = true
	app.confirmMsg = "Test confirm?"

	view := app.View()

	if len(view) == 0 {
		t.Error("View() in confirm mode returned empty string")
	}
}

func TestApp_View_FullscreenLogMode(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.fullscreenLog = true

	view := app.View()

	if len(view) == 0 {
		t.Error("View() in fullscreen log mode returned empty string")
	}
}

func TestApp_Update_WindowSizeMsg(t *testing.T) {
	app := New()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
	}
}

func TestApp_Update_WorkflowsLoadedMsg(t *testing.T) {
	app := New()

	workflows := []github.Workflow{
		{ID: 1, Name: "CI"},
		{ID: 2, Name: "Deploy"},
	}
	msg := WorkflowsLoadedMsg{Workflows: workflows}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.loading {
		t.Error("loading should be false after WorkflowsLoadedMsg")
	}
	if updated.workflows.Len() != 2 {
		t.Errorf("workflows.Len() = %d, want 2", updated.workflows.Len())
	}
}

func TestApp_Update_WorkflowsLoadedMsg_WithError(t *testing.T) {
	app := New()

	msg := WorkflowsLoadedMsg{Err: errors.New("not found")}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.err == nil {
		t.Error("err should be set after WorkflowsLoadedMsg with error")
	}
}

func TestApp_Update_RunsLoadedMsg(t *testing.T) {
	app := New()

	runs := []github.Run{
		{ID: 1, Name: "Run 1"},
	}
	msg := RunsLoadedMsg{Runs: runs}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.runs.Len() != 1 {
		t.Errorf("runs.Len() = %d, want 1", updated.runs.Len())
	}
}

func TestApp_Update_JobsLoadedMsg(t *testing.T) {
	app := New()

	jobs := []github.Job{
		{ID: 1, Name: "build"},
	}
	msg := JobsLoadedMsg{Jobs: jobs}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.jobs.Len() != 1 {
		t.Errorf("jobs.Len() = %d, want 1", updated.jobs.Len())
	}
}

func TestApp_Update_LogsLoadedMsg(t *testing.T) {
	app := New()
	app.width = 80
	app.height = 20

	// Set up jobs first so LogsLoadedMsg can match the selected job
	app.jobs.SetItems([]github.Job{{ID: 100, Name: "test"}})

	msg := LogsLoadedMsg{JobID: 100, Logs: "test logs"}
	model, _ := app.Update(msg)
	updated := model.(*App)

	// Just ensure no panic
	_ = updated.View()
}

func TestApp_Update_FlashClearMsg(t *testing.T) {
	app := New()
	app.flashMsg = "Test flash"

	msg := FlashClearMsg{}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.flashMsg != "" {
		t.Errorf("flashMsg = %q, want empty", updated.flashMsg)
	}
}

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

func TestApp_HandleMouseEvent_WheelUp(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.workflows.SetItems([]github.Workflow{
		{ID: 1, Name: "CI"},
		{ID: 2, Name: "Deploy"},
	})
	app.workflows.SelectNext() // Move to index 1

	if app.workflows.SelectedIndex() != 1 {
		t.Fatalf("Setup failed: SelectedIndex = %d, want 1", app.workflows.SelectedIndex())
	}

	msg := tea.MouseMsg{Button: tea.MouseButtonWheelUp}
	app.handleMouseEvent(msg)

	if app.workflows.SelectedIndex() != 0 {
		t.Errorf("After wheel up: SelectedIndex = %d, want 0", app.workflows.SelectedIndex())
	}
}

func TestApp_HandleMouseEvent_WheelDown(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.workflows.SetItems([]github.Workflow{
		{ID: 1, Name: "CI"},
		{ID: 2, Name: "Deploy"},
	})

	if app.workflows.SelectedIndex() != 0 {
		t.Fatalf("Setup failed: SelectedIndex = %d, want 0", app.workflows.SelectedIndex())
	}

	msg := tea.MouseMsg{Button: tea.MouseButtonWheelDown}
	app.handleMouseEvent(msg)

	if app.workflows.SelectedIndex() != 1 {
		t.Errorf("After wheel down: SelectedIndex = %d, want 1", app.workflows.SelectedIndex())
	}
}

func TestApp_HandleMouseEvent_IgnoredWhenPopupShown(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.showHelp = true

	msg := tea.MouseMsg{Button: tea.MouseButtonWheelDown}
	app.handleMouseEvent(msg)

	// Mouse events should be ignored when popup is shown
	// No crash = test passes
}

func TestApp_HandleClick_WorkflowsPane(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.workflows.SetItems([]github.Workflow{
		{ID: 1, Name: "CI"},
		{ID: 2, Name: "Deploy"},
		{ID: 3, Name: "Test"},
	})
	app.focusedPane = RunsPane // Start in Runs pane

	// Click on Workflows panel (y=3 should be around item index 1)
	app.handleClick(10, 3)

	if app.focusedPane != WorkflowsPane {
		t.Errorf("After click: focusedPane = %v, want WorkflowsPane", app.focusedPane)
	}
}

func TestApp_HandleClick_OutsideLeftPanel(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.focusedPane = WorkflowsPane

	// Click on right panel (x=50 is outside left sidebar which is ~36 wide at 30%)
	app.handleClick(50, 10)

	// Should remain unchanged
	if app.focusedPane != WorkflowsPane {
		t.Errorf("Click on right panel should not change focus: got %v", app.focusedPane)
	}
}

func TestApp_HandleMouseEvent_ScrollInRunsPane(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.focusedPane = RunsPane
	app.runs.SetItems([]github.Run{
		{ID: 1, Name: "Run 1"},
		{ID: 2, Name: "Run 2"},
	})

	msg := tea.MouseMsg{Button: tea.MouseButtonWheelDown}
	app.handleMouseEvent(msg)

	if app.runs.SelectedIndex() != 1 {
		t.Errorf("After wheel down in RunsPane: SelectedIndex = %d, want 1", app.runs.SelectedIndex())
	}

	msg = tea.MouseMsg{Button: tea.MouseButtonWheelUp}
	app.handleMouseEvent(msg)

	if app.runs.SelectedIndex() != 0 {
		t.Errorf("After wheel up in RunsPane: SelectedIndex = %d, want 0", app.runs.SelectedIndex())
	}
}

func TestApp_HandleMouseEvent_ScrollInJobsPane(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.focusedPane = JobsPane
	app.jobs.SetItems([]github.Job{
		{ID: 1, Name: "build"},
		{ID: 2, Name: "test"},
	})

	msg := tea.MouseMsg{Button: tea.MouseButtonWheelDown}
	app.handleMouseEvent(msg)

	if app.jobs.SelectedIndex() != 1 {
		t.Errorf("After wheel down in JobsPane: SelectedIndex = %d, want 1", app.jobs.SelectedIndex())
	}

	msg = tea.MouseMsg{Button: tea.MouseButtonWheelUp}
	app.handleMouseEvent(msg)

	if app.jobs.SelectedIndex() != 0 {
		t.Errorf("After wheel up in JobsPane: SelectedIndex = %d, want 0", app.jobs.SelectedIndex())
	}
}

func TestApp_HandleClick_RunsPane(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.runs.SetItems([]github.Run{
		{ID: 1, Name: "Run 1"},
		{ID: 2, Name: "Run 2"},
	})
	app.focusedPane = WorkflowsPane

	// Click on Runs panel (y should be in the Runs panel area)
	panelHeight := (app.height - 1) / 3
	app.handleClick(10, panelHeight+3)

	if app.focusedPane != RunsPane {
		t.Errorf("After click: focusedPane = %v, want RunsPane", app.focusedPane)
	}
}

func TestApp_HandleClick_JobsPane(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.jobs.SetItems([]github.Job{
		{ID: 1, Name: "build"},
		{ID: 2, Name: "test"},
	})
	app.focusedPane = WorkflowsPane

	// Click on Jobs panel (y should be in the Jobs panel area)
	panelHeight := (app.height - 1) / 3
	app.handleClick(10, 2*panelHeight+3)

	if app.focusedPane != JobsPane {
		t.Errorf("After click: focusedPane = %v, want JobsPane", app.focusedPane)
	}
}

func TestApp_HandleMouseEvent_LeftClickRelease(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.workflows.SetItems([]github.Workflow{
		{ID: 1, Name: "CI"},
		{ID: 2, Name: "Deploy"},
	})
	app.focusedPane = RunsPane

	// Test left click release triggers handleClick
	msg := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionRelease,
		X:      10,
		Y:      3,
	}
	app.handleMouseEvent(msg)

	// Should have switched to Workflows pane
	if app.focusedPane != WorkflowsPane {
		t.Errorf("After left click release: focusedPane = %v, want WorkflowsPane", app.focusedPane)
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

func TestApp_ConfirmCancelRun_NoSelection(t *testing.T) {
	app := New()
	app.focusedPane = RunsPane

	cmd := app.confirmCancelRun()

	if cmd != nil {
		t.Error("confirmCancelRun with no selection should return nil")
	}
}

func TestApp_ConfirmCancelRun_NotRunning(t *testing.T) {
	app := New()
	app.focusedPane = RunsPane
	app.runs.SetItems([]github.Run{
		{ID: 1, Status: "completed", Conclusion: "success"},
	})

	cmd := app.confirmCancelRun()

	if cmd != nil {
		t.Error("confirmCancelRun for non-running run should return nil")
	}
}

func TestApp_ConfirmCancelRun_Running(t *testing.T) {
	app := New()
	app.focusedPane = RunsPane
	app.runs.SetItems([]github.Run{
		{ID: 1, Status: "in_progress"},
	})

	app.confirmCancelRun()

	if !app.showConfirm {
		t.Error("confirmCancelRun should show confirm dialog")
	}
	if app.confirmFn == nil {
		t.Error("confirmCancelRun should set confirmFn")
	}
}

func TestApp_RerunWorkflow_NoSelection(t *testing.T) {
	app := New()

	cmd := app.rerunWorkflow()

	if cmd != nil {
		t.Error("rerunWorkflow with no selection should return nil")
	}
}

func TestApp_RerunFailedJobs_NotFailed(t *testing.T) {
	app := New()
	app.runs.SetItems([]github.Run{
		{ID: 1, Status: "completed", Conclusion: "success"},
	})

	cmd := app.rerunFailedJobs()

	if cmd != nil {
		t.Error("rerunFailedJobs for non-failed run should return nil")
	}
}

func TestApp_RefreshCurrentWorkflow_NoSelection(t *testing.T) {
	app := New()

	cmd := app.refreshCurrentWorkflow()

	if cmd != nil {
		t.Error("refreshCurrentWorkflow with no selection should return nil")
	}
}

func TestFormatRunNumber(t *testing.T) {
	tests := []struct {
		id   int64
		want string
	}{
		{1, "1"},
		{123, "123"},
		{1234567890, "1234567890"},
	}

	for _, tt := range tests {
		got := formatRunNumber(tt.id)
		if got != tt.want {
			t.Errorf("formatRunNumber(%d) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestApp_PaneWidths(t *testing.T) {
	app := New()
	app.width = 100
	app.height = 50

	workflowsWidth := app.workflowsPaneWidth()
	runsWidth := app.runsPaneWidth()
	logsWidth := app.logPaneWidth()

	if workflowsWidth <= 0 {
		t.Errorf("workflowsPaneWidth = %d, should be positive", workflowsWidth)
	}
	if runsWidth <= 0 {
		t.Errorf("runsPaneWidth = %d, should be positive", runsWidth)
	}
	if logsWidth <= 0 {
		t.Errorf("logPaneWidth = %d, should be positive", logsWidth)
	}

	// They should roughly add up to the total width
	total := workflowsWidth + runsWidth + logsWidth
	if total > 100 {
		t.Errorf("total width %d exceeds app width 100", total)
	}
}

func TestApp_PaneHeight(t *testing.T) {
	app := New()
	app.height = 50

	height := app.paneHeight()
	if height != 48 {
		t.Errorf("paneHeight = %d, want 48", height)
	}

	logHeight := app.logPaneHeight()
	if logHeight != 48 {
		t.Errorf("logPaneHeight = %d, want 48", logHeight)
	}
}

func TestApp_RenderPanes(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40

	// Just test they don't panic
	_ = app.buildWorkflowsPanel(30, 10)
	_ = app.buildRunsPanel(30, 10)
	_ = app.buildJobsPanel(30, 10)
	_ = app.buildDetailPanel(60, 30)
	_ = app.renderStatusBar()
	_ = app.renderHelp()

	app.confirmMsg = "Test?"
	_ = app.renderConfirmDialog()

	_ = app.renderFullscreenLog()
}

func TestApp_RenderStatusBar_States(t *testing.T) {
	app := New()
	app.width = 100
	app.height = 40

	// Normal state
	bar := app.renderStatusBar()
	if len(bar) == 0 {
		t.Error("renderStatusBar returned empty string")
	}

	// Filtering state
	app.filtering = true
	bar = app.renderStatusBar()
	if len(bar) == 0 {
		t.Error("renderStatusBar in filtering mode returned empty string")
	}
	app.filtering = false

	// Flash message
	app.flashMsg = "Test flash"
	bar = app.renderStatusBar()
	if len(bar) == 0 {
		t.Error("renderStatusBar with flash message returned empty string")
	}
	app.flashMsg = ""

	// Error state
	app.err = errors.New("not found")
	bar = app.renderStatusBar()
	if len(bar) == 0 {
		t.Error("renderStatusBar with error returned empty string")
	}
}

func TestApp_StartLogPolling_NoClient(t *testing.T) {
	// This test is skipped because StartLogPolling requires adaptivePoller to be set
	// which requires a client
	t.Skip("StartLogPolling requires adaptivePoller to be set")
}

func TestApp_StopLogPolling(t *testing.T) {
	app := New()

	// Should not panic when logPoller is nil
	app.StopLogPolling()

	if app.logPoller != nil {
		t.Error("logPoller should remain nil")
	}
}

func TestApp_OnRunSelectionChange(t *testing.T) {
	mock := newMockClient(nil)
	app := New(WithClient(mock))
	app.runs.SetItems([]github.Run{
		{ID: 1, Name: "Run 1"},
	})

	cmd := app.onRunSelectionChange()
	if cmd == nil {
		t.Error("onRunSelectionChange should return a command when run is selected")
	}
}

func TestApp_OnJobSelectionChange(t *testing.T) {
	mock := newMockClient(nil)
	app := New(WithClient(mock))
	app.jobs.SetItems([]github.Job{
		{ID: 1, Name: "build"},
	})

	cmd := app.onJobSelectionChange()
	if cmd == nil {
		t.Error("onJobSelectionChange should return a command when job is selected")
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

func TestApp_FetchCmds_WithClient(t *testing.T) {
	mock := newMockClient(nil)
	app := New(WithClient(mock))

	// These should return commands when client is set
	if cmd := app.fetchWorkflowsCmd(); cmd == nil {
		t.Error("fetchWorkflowsCmd should return command when client is set")
	}
	if cmd := app.fetchRunsCmd(1); cmd == nil {
		t.Error("fetchRunsCmd should return command when client is set")
	}
	if cmd := app.fetchJobsCmd(1); cmd == nil {
		t.Error("fetchJobsCmd should return command when client is set")
	}
	if cmd := app.fetchLogsCmd(1); cmd == nil {
		t.Error("fetchLogsCmd should return command when client is set")
	}
}

func TestApp_RerunWorkflow_WithSelection(t *testing.T) {
	mock := newMockClient(nil)
	app := New(WithClient(mock))
	app.runs.SetItems([]github.Run{
		{ID: 1, Name: "Run 1"},
	})

	cmd := app.rerunWorkflow()
	if cmd == nil {
		t.Error("rerunWorkflow should return command when run is selected")
	}
}

func TestApp_RerunFailedJobs_WithFailedRun(t *testing.T) {
	mock := newMockClient(nil)
	app := New(WithClient(mock))
	app.runs.SetItems([]github.Run{
		{ID: 1, Status: "completed", Conclusion: "failure"},
	})

	cmd := app.rerunFailedJobs()
	if cmd == nil {
		t.Error("rerunFailedJobs should return command when run is failed")
	}
}

func TestApp_RefreshCurrentWorkflow_WithSelection(t *testing.T) {
	mock := newMockClient(nil)
	app := New(WithClient(mock))
	app.workflows.SetItems([]github.Workflow{
		{ID: 1, Name: "CI"},
	})

	cmd := app.refreshCurrentWorkflow()
	if cmd == nil {
		t.Error("refreshCurrentWorkflow should return command when workflow is selected")
	}
}

func TestApp_HandleDetailPanelClick_StepSelection(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.detailTab = LogsTab
	app.jobs.SetItems([]github.Job{{ID: 1, Name: "build"}})

	// Set up parsed logs with steps
	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Step 1
2024-01-15T10:00:01.000Z Line 1
2024-01-15T10:00:02.000Z ##[endgroup]
2024-01-15T10:00:03.000Z ##[group]Step 2
2024-01-15T10:00:04.000Z Line 2
2024-01-15T10:00:05.000Z ##[endgroup]`
	app.parsedLogs = ParseLogs(rawLogs)

	// Initial state: selectedStepIdx should be -1 (All logs)
	if app.selectedStepIdx != -1 {
		t.Fatalf("Initial selectedStepIdx = %d, want -1", app.selectedStepIdx)
	}

	leftWidth := int(float64(app.width) * 0.30)

	// Click on "All logs" (y=5)
	app.handleDetailPanelClick(leftWidth+10, 5, leftWidth, app.height-1)
	if app.selectedStepIdx != -1 {
		t.Errorf("After click on All logs: selectedStepIdx = %d, want -1", app.selectedStepIdx)
	}

	// Click on Step 1 (y=6)
	app.handleDetailPanelClick(leftWidth+10, 6, leftWidth, app.height-1)
	if app.selectedStepIdx != 0 {
		t.Errorf("After click on Step 1: selectedStepIdx = %d, want 0", app.selectedStepIdx)
	}

	// Click on Step 2 (y=7)
	app.handleDetailPanelClick(leftWidth+10, 7, leftWidth, app.height-1)
	if app.selectedStepIdx != 1 {
		t.Errorf("After click on Step 2: selectedStepIdx = %d, want 1", app.selectedStepIdx)
	}
}

func TestApp_HandleDetailPanelClick_NoSteps(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.detailTab = LogsTab
	app.parsedLogs = nil // No parsed logs

	leftWidth := int(float64(app.width) * 0.30)

	// Click should not cause panic when there are no steps
	app.handleDetailPanelClick(leftWidth+10, 5, leftWidth, app.height-1)
	// No panic = test passes
}

func TestApp_HandleDetailPanelClick_InfoTab(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.detailTab = InfoTab // Not LogsTab

	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Step 1
2024-01-15T10:00:01.000Z ##[endgroup]`
	app.parsedLogs = ParseLogs(rawLogs)

	leftWidth := int(float64(app.width) * 0.30)

	// Click should be ignored in Info tab
	app.handleDetailPanelClick(leftWidth+10, 5, leftWidth, app.height-1)
	// selectedStepIdx should remain unchanged
	if app.selectedStepIdx != -1 {
		t.Errorf("Click in InfoTab should not change selectedStepIdx")
	}
}

func TestApp_HandleScrollInDetailPanel_StepList(t *testing.T) {
	app := New()
	app.width = 120
	app.height = 40
	app.detailTab = LogsTab
	app.stepListFocused = true
	app.jobs.SetItems([]github.Job{{ID: 1, Name: "build"}})

	rawLogs := `2024-01-15T10:00:00.000Z ##[group]Step 1
2024-01-15T10:00:01.000Z ##[endgroup]
2024-01-15T10:00:02.000Z ##[group]Step 2
2024-01-15T10:00:03.000Z ##[endgroup]`
	app.parsedLogs = ParseLogs(rawLogs)

	leftWidth := int(float64(app.width) * 0.30)
	app.mouseX = leftWidth + 10 // Mouse in detail panel

	// Initial state
	if app.selectedStepIdx != -1 {
		t.Fatalf("Initial selectedStepIdx = %d, want -1", app.selectedStepIdx)
	}

	// Scroll down should select Step 1
	app.handleScrollDown()
	if app.selectedStepIdx != 0 {
		t.Errorf("After scroll down: selectedStepIdx = %d, want 0", app.selectedStepIdx)
	}

	// Scroll down should select Step 2
	app.handleScrollDown()
	if app.selectedStepIdx != 1 {
		t.Errorf("After second scroll down: selectedStepIdx = %d, want 1", app.selectedStepIdx)
	}

	// Scroll up should select Step 1
	app.handleScrollUp()
	if app.selectedStepIdx != 0 {
		t.Errorf("After scroll up: selectedStepIdx = %d, want 0", app.selectedStepIdx)
	}

	// Scroll up should select All logs
	app.handleScrollUp()
	if app.selectedStepIdx != -1 {
		t.Errorf("After second scroll up: selectedStepIdx = %d, want -1", app.selectedStepIdx)
	}
}
