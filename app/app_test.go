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
	mock := &github.MockClient{}
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

	msg := LogsLoadedMsg{Logs: "test logs"}
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

	// Navigate down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	app.handleKeyPress(msg)

	if app.workflows.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", app.workflows.SelectedIndex())
	}

	// Navigate up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
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

	// Move right to LogsPane
	app.handleKeyPress(msg)

	if app.focusedPane != LogsPane {
		t.Errorf("focusedPane = %v, want LogsPane", app.focusedPane)
	}

	// Move left back to RunsPane
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	app.handleKeyPress(msg)

	if app.focusedPane != RunsPane {
		t.Errorf("focusedPane = %v, want RunsPane", app.focusedPane)
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
	app.focusedPane = LogsPane

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}
	app.handleKeyPress(msg)

	if !app.fullscreenLog {
		t.Error("L key in LogsPane should enable fullscreenLog")
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
		{LogsPane, RunsPane},
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
		{RunsPane, LogsPane},
		{LogsPane, LogsPane}, // Can't go past last
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
	_ = app.renderWorkflowsPane()
	_ = app.renderRunsPane()
	_ = app.renderLogsPane()
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
	mock := &github.MockClient{}
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
	mock := &github.MockClient{}
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
	mock := &github.MockClient{}
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

func TestApp_NavigateUpInLogsPane(t *testing.T) {
	mock := &github.MockClient{}
	app := New(WithClient(mock))
	app.focusedPane = LogsPane
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

func TestApp_ApplyFilterOnLogsPane(t *testing.T) {
	app := New()
	app.focusedPane = LogsPane
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
	mock := &github.MockClient{}
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
	mock := &github.MockClient{}
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
	mock := &github.MockClient{}
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
	mock := &github.MockClient{}
	app := New(WithClient(mock))
	app.workflows.SetItems([]github.Workflow{
		{ID: 1, Name: "CI"},
	})

	cmd := app.refreshCurrentWorkflow()
	if cmd == nil {
		t.Error("refreshCurrentWorkflow should return command when workflow is selected")
	}
}
