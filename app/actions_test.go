package app

import (
	"testing"

	"github.com/nnnkkk7/lazyactions/github"
)

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
