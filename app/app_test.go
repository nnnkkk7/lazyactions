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

	if updated.workflows.Len() != 2 {
		t.Errorf("workflows.Len() = %d, want 2", updated.workflows.Len())
	}
	if updated.loading {
		t.Error("loading should be false after workflows loaded")
	}
}

func TestApp_Update_WorkflowsLoadedMsg_WithError(t *testing.T) {
	app := New()

	msg := WorkflowsLoadedMsg{Err: errAPI}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.err == nil {
		t.Error("err should be set when WorkflowsLoadedMsg has error")
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
		{ID: 1, Name: "build", Status: "completed"},
	}
	msg := JobsLoadedMsg{Jobs: jobs}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.jobs.Len() != 1 {
		t.Errorf("jobs.Len() = %d, want 1", updated.jobs.Len())
	}
}

func TestApp_Update_JobsLoadedMsg_QueuedJob(t *testing.T) {
	app := New()

	jobs := []github.Job{
		{ID: 1, Name: "build", Status: "queued"},
	}
	msg := JobsLoadedMsg{Jobs: jobs}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.jobs.Len() != 1 {
		t.Errorf("jobs.Len() = %d, want 1", updated.jobs.Len())
	}
	// parsedLogs should be nil for queued job (no logs fetched)
	if updated.parsedLogs != nil {
		t.Error("parsedLogs should be nil for queued job")
	}
}

func TestApp_Update_JobsLoadedMsg_InProgressJob(t *testing.T) {
	app := New()

	jobs := []github.Job{
		{ID: 1, Name: "build", Status: "in_progress"},
	}
	msg := JobsLoadedMsg{Jobs: jobs}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.jobs.Len() != 1 {
		t.Errorf("jobs.Len() = %d, want 1", updated.jobs.Len())
	}
	// parsedLogs should be nil for in_progress job (no logs fetched)
	if updated.parsedLogs != nil {
		t.Error("parsedLogs should be nil for in_progress job")
	}
}

func TestApp_Update_LogsLoadedMsg(t *testing.T) {
	app := New()
	app.jobs.SetItems([]github.Job{{ID: 1, Name: "build", Status: "completed"}})

	msg := LogsLoadedMsg{JobID: 1, Logs: "test logs"}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.parsedLogs == nil {
		t.Error("parsedLogs should be set after logs loaded")
	}
}

func TestApp_Update_LogsLoadedMsg_ErrorForIncompleteJob(t *testing.T) {
	app := New()
	app.jobs.SetItems([]github.Job{{ID: 1, Name: "build", Status: "in_progress"}})

	msg := LogsLoadedMsg{JobID: 1, Logs: "", Err: errors.New("not found")}
	model, _ := app.Update(msg)
	updated := model.(*App)

	// Should not set a.err for incomplete job
	if updated.err != nil {
		t.Error("err should be nil for incomplete job")
	}
}

func TestApp_Update_LogsLoadedMsg_ErrorForCompletedJob(t *testing.T) {
	app := New()
	app.jobs.SetItems([]github.Job{{ID: 1, Name: "build", Status: "completed"}})

	msg := LogsLoadedMsg{JobID: 1, Logs: "", Err: errors.New("not found")}
	model, _ := app.Update(msg)
	updated := model.(*App)

	// Should not set a.err (we changed this behavior)
	if updated.err != nil {
		t.Error("err should be nil even for completed job error")
	}
}

func TestApp_Update_FlashClearMsg(t *testing.T) {
	app := New()
	app.flashMsg = "Test message"

	msg := FlashClearMsg{}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if updated.flashMsg != "" {
		t.Errorf("flashMsg = %q, want empty string", updated.flashMsg)
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
