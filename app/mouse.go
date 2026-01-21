package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleMouseEvent handles mouse events
func (a *App) handleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Always track mouse position for hover highlighting
	a.mouseX = msg.X
	a.mouseY = msg.Y

	// Ignore actions when popups are shown
	if a.showHelp || a.showConfirm || a.fullscreenLog || a.filtering {
		return a, nil
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		return a.handleScrollUp()
	case tea.MouseButtonWheelDown:
		return a.handleScrollDown()
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionRelease {
			return a.handleClick(msg.X, msg.Y)
		}
	}
	return a, nil
}

// handleClick handles mouse click events
func (a *App) handleClick(x, y int) (tea.Model, tea.Cmd) {
	leftWidth := a.leftPanelWidth()
	totalHeight, panelHeight := a.panelLayout()

	// Handle clicks in the right panel (detail view)
	if x >= leftWidth {
		return a.handleDetailPanelClick(x, y, leftWidth, totalHeight)
	}

	// Determine which panel was clicked (left sidebar)
	if y < panelHeight {
		// Workflows panel
		a.focusedPane = WorkflowsPane
		itemIdx := y - BorderOffset
		if itemIdx >= 0 && itemIdx < a.workflows.Len() {
			a.workflows.Select(itemIdx)
			return a, a.onWorkflowSelectionChange()
		}
	} else if y < 2*panelHeight {
		// Runs panel
		a.focusedPane = RunsPane
		itemIdx := y - panelHeight - BorderOffset
		if itemIdx >= 0 && itemIdx < a.runs.Len() {
			a.runs.Select(itemIdx)
			return a, a.onRunSelectionChange()
		}
	} else if y < totalHeight {
		// Jobs panel
		a.focusedPane = JobsPane
		itemIdx := y - 2*panelHeight - BorderOffset
		if itemIdx >= 0 && itemIdx < a.jobs.Len() {
			a.jobs.Select(itemIdx)
			return a, a.onJobSelectionChange()
		}
	}

	return a, nil
}

// handleDetailPanelClick handles mouse clicks in the detail panel (right side)
func (a *App) handleDetailPanelClick(_, y, _, _ int) (tea.Model, tea.Cmd) {
	// Only handle clicks in Logs tab with step list
	if a.detailTab != LogsTab || a.parsedLogs == nil || len(a.parsedLogs.Steps) == 0 {
		return a, nil
	}

	// Calculate the step list area in the detail panel
	// Layout of buildLogsContent:
	// Line 0 (y=1): "  Logs: job_name" (title)
	// Line 1 (y=2): separator
	// Line 2 (y=3): "  Steps: (hint)"
	// Line 3 (y=4): empty
	// Line 4 (y=5): "All logs" option (selectedStepIdx = -1)
	// Line 5+ (y=6+): individual steps (selectedStepIdx = 0, 1, 2, ...)

	// Content starts at y=1 (after top border)
	// Step list starts at content line 4 (y=5)
	stepListStartY := 5 // "All logs" is at y=5
	stepCount := len(a.parsedLogs.Steps)

	// Check if click is in the step list area
	if y >= stepListStartY && y < stepListStartY+1+stepCount {
		clickedIdx := y - stepListStartY - 1 // -1 because "All logs" is at index -1

		// Validate the clicked index
		if clickedIdx >= -1 && clickedIdx < stepCount {
			a.selectedStepIdx = clickedIdx
			a.stepListFocused = true
			a.updateLogViewContent()
			a.logView.GotoTop()
		}
	}

	return a, nil
}

// handleScrollUp handles mouse wheel up
func (a *App) handleScrollUp() (tea.Model, tea.Cmd) {
	// Calculate left panel width for determining scroll context
	leftWidth := a.leftPanelWidth()

	// If mouse is in the detail panel and we're in Logs tab with steps, scroll the step list
	if a.mouseX >= leftWidth && a.detailTab == LogsTab && a.parsedLogs != nil && len(a.parsedLogs.Steps) > 0 {
		if a.stepListFocused {
			a.navigateStepUp()
		} else {
			a.logView.ScrollUp()
		}
		return a, nil
	}

	// Otherwise, scroll the focused left panel
	switch a.focusedPane {
	case WorkflowsPane:
		a.workflows.SelectPrev()
		return a, a.onWorkflowSelectionChange()
	case RunsPane:
		a.runs.SelectPrev()
		return a, a.onRunSelectionChange()
	case JobsPane:
		a.jobs.SelectPrev()
		return a, a.onJobSelectionChange()
	}
	return a, nil
}

// handleScrollDown handles mouse wheel down
func (a *App) handleScrollDown() (tea.Model, tea.Cmd) {
	// Calculate left panel width for determining scroll context
	leftWidth := a.leftPanelWidth()

	// If mouse is in the detail panel and we're in Logs tab with steps, scroll the step list
	if a.mouseX >= leftWidth && a.detailTab == LogsTab && a.parsedLogs != nil && len(a.parsedLogs.Steps) > 0 {
		if a.stepListFocused {
			a.navigateStepDown()
		} else {
			a.logView.ScrollDown()
		}
		return a, nil
	}

	// Otherwise, scroll the focused left panel
	switch a.focusedPane {
	case WorkflowsPane:
		a.workflows.SelectNext()
		return a, a.onWorkflowSelectionChange()
	case RunsPane:
		a.runs.SelectNext()
		return a, a.onRunSelectionChange()
	case JobsPane:
		a.jobs.SelectNext()
		return a, a.onJobSelectionChange()
	}
	return a, nil
}

// onWorkflowSelectionChange handles workflow selection change
func (a *App) onWorkflowSelectionChange() tea.Cmd {
	if wf, ok := a.workflows.Selected(); ok {
		a.loading = true
		return a.fetchRunsCmd(wf.ID)
	}
	return nil
}

// onRunSelectionChange handles run selection change
func (a *App) onRunSelectionChange() tea.Cmd {
	if run, ok := a.runs.Selected(); ok {
		a.loading = true
		return a.fetchJobsCmd(run.ID)
	}
	return nil
}

// onJobSelectionChange handles job selection change
func (a *App) onJobSelectionChange() tea.Cmd {
	if job, ok := a.jobs.Selected(); ok {
		// Clear current logs and show loading state while fetching new logs
		a.logView.SetContent("Loading logs...")
		// Reset step selection for new job
		a.parsedLogs = nil
		a.selectedStepIdx = -1
		a.stepListFocused = true
		return a.fetchLogsCmd(job.ID)
	}
	return nil
}

// updateLogViewContent updates the log view with the currently selected step's logs
func (a *App) updateLogViewContent() {
	if a.parsedLogs == nil {
		a.logView.SetContent("No logs available")
		return
	}

	// Get logs for the selected step (formatted with syntax highlighting)
	logs := a.parsedLogs.FormatStepLogsWithColor(a.selectedStepIdx)
	if logs == "" {
		logs = "No logs available"
	}

	// Wrap log lines to fit within viewport width
	wrappedLogs := wrapLines(logs, a.logPaneWidth()-4)
	a.logView.SetContent(wrappedLogs)
}

// navigateStepUp moves step selection up
func (a *App) navigateStepUp() {
	if a.parsedLogs == nil || len(a.parsedLogs.Steps) == 0 {
		return
	}
	// -1 is "All logs", 0 to len-1 are specific steps
	if a.selectedStepIdx > -1 {
		a.selectedStepIdx--
		a.updateLogViewContent()
		a.logView.GotoTop()
	}
}

// navigateStepDown moves step selection down
func (a *App) navigateStepDown() {
	if a.parsedLogs == nil || len(a.parsedLogs.Steps) == 0 {
		return
	}
	maxIdx := len(a.parsedLogs.Steps) - 1
	if a.selectedStepIdx < maxIdx {
		a.selectedStepIdx++
		a.updateLogViewContent()
		a.logView.GotoTop()
	}
}
