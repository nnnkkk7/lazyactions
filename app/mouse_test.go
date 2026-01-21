package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/lazyactions/github"
)

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
