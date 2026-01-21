package app

import (
	"errors"
	"testing"
)

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
