package app

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyPress handles key press events
func (a *App) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	// Handle filter input mode
	if a.filtering {
		return a.handleFilterInput(msg)
	}

	// Handle confirm dialog
	if a.showConfirm {
		return a.handleConfirmInput(msg)
	}

	switch {
	case key.Matches(msg, a.keys.Quit):
		return tea.Quit

	case key.Matches(msg, a.keys.Help):
		a.showHelp = !a.showHelp

	case key.Matches(msg, a.keys.Escape):
		if a.showHelp {
			a.showHelp = false
		} else if a.fullscreenLog {
			a.fullscreenLog = false
		} else if a.detailTab == LogsTab && a.focusedPane == JobsPane && !a.stepListFocused {
			// Return focus to step list from log content
			a.stepListFocused = true
		} else if a.err != nil {
			a.err = nil
			return a.refreshAll()
		}

	case key.Matches(msg, a.keys.Enter):
		// When in Logs tab with step list focused, Enter focuses on log content
		if a.detailTab == LogsTab && a.focusedPane == JobsPane && a.stepListFocused {
			a.stepListFocused = false
		}

	case key.Matches(msg, a.keys.Up):
		return a.navigateUp()

	case key.Matches(msg, a.keys.Down):
		return a.navigateDown()

	case key.Matches(msg, a.keys.PanelUp):
		return a.focusPrevPaneWithSelect()

	case key.Matches(msg, a.keys.PanelDown):
		return a.focusNextPaneWithSelect()

	case key.Matches(msg, a.keys.Left), key.Matches(msg, a.keys.ShiftTab):
		a.focusPrevPane()

	case key.Matches(msg, a.keys.Right), key.Matches(msg, a.keys.Tab):
		a.focusNextPane()

	case key.Matches(msg, a.keys.Filter):
		a.filtering = true
		a.filterInput.Focus()

	case key.Matches(msg, a.keys.FullLog):
		if a.focusedPane == JobsPane {
			a.fullscreenLog = true
		}

	case key.Matches(msg, a.keys.Cancel):
		if a.focusedPane == RunsPane {
			return a.confirmCancelRun()
		}

	case key.Matches(msg, a.keys.Rerun):
		if a.focusedPane == RunsPane {
			return a.rerunWorkflow()
		}

	case key.Matches(msg, a.keys.RerunFailed):
		if a.focusedPane == RunsPane {
			return a.rerunFailedJobs()
		}

	case key.Matches(msg, a.keys.Trigger):
		if a.focusedPane == WorkflowsPane {
			return a.triggerWorkflow()
		}

	case key.Matches(msg, a.keys.Yank):
		return a.yankURL()

	case key.Matches(msg, a.keys.Refresh):
		return a.refreshAll()

	case key.Matches(msg, a.keys.InfoTab):
		a.detailTab = InfoTab

	case key.Matches(msg, a.keys.LogsTab):
		a.detailTab = LogsTab
	}

	return nil
}

// handleFilterInput handles input when in filter mode
func (a *App) handleFilterInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		a.filtering = false
		a.filterInput.Blur()
		a.applyFilter("")
	case "enter":
		a.filtering = false
		a.filterInput.Blur()
		a.applyFilter(a.filterInput.Value())
	default:
		var cmd tea.Cmd
		a.filterInput, cmd = a.filterInput.Update(msg)
		return cmd
	}
	return nil
}

// handleConfirmInput handles input when in confirm dialog
func (a *App) handleConfirmInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		a.showConfirm = false
		if a.confirmFn != nil {
			return a.confirmFn()
		}
	case "n", "N", "esc":
		a.showConfirm = false
		a.confirmFn = nil
	}
	return nil
}

// applyFilter applies filter to the currently focused pane
func (a *App) applyFilter(filter string) {
	switch a.focusedPane {
	case WorkflowsPane:
		a.workflows.SetFilter(filter)
	case RunsPane:
		a.runs.SetFilter(filter)
	case JobsPane:
		a.jobs.SetFilter(filter)
	}
}

// navigateUp moves selection up in the current pane
func (a *App) navigateUp() tea.Cmd {
	switch a.focusedPane {
	case WorkflowsPane:
		a.workflows.SelectPrev()
		return a.onWorkflowSelectionChange()
	case RunsPane:
		a.runs.SelectPrev()
		return a.onRunSelectionChange()
	case JobsPane:
		// If in Logs tab and step list is focused, navigate steps
		if a.detailTab == LogsTab && a.stepListFocused && a.parsedLogs != nil && len(a.parsedLogs.Steps) > 0 {
			a.navigateStepUp()
			return nil
		}
		// If not step focused, scroll log content
		if a.detailTab == LogsTab && !a.stepListFocused {
			a.logView.ScrollUp()
			return nil
		}
		a.jobs.SelectPrev()
		return a.onJobSelectionChange()
	}
	return nil
}

// navigateDown moves selection down in the current pane
func (a *App) navigateDown() tea.Cmd {
	switch a.focusedPane {
	case WorkflowsPane:
		a.workflows.SelectNext()
		return a.onWorkflowSelectionChange()
	case RunsPane:
		a.runs.SelectNext()
		return a.onRunSelectionChange()
	case JobsPane:
		// If in Logs tab and step list is focused, navigate steps
		if a.detailTab == LogsTab && a.stepListFocused && a.parsedLogs != nil && len(a.parsedLogs.Steps) > 0 {
			a.navigateStepDown()
			return nil
		}
		// If not step focused, scroll log content
		if a.detailTab == LogsTab && !a.stepListFocused {
			a.logView.ScrollDown()
			return nil
		}
		a.jobs.SelectNext()
		return a.onJobSelectionChange()
	}
	return nil
}

// focusPrevPane moves focus to the previous pane
func (a *App) focusPrevPane() {
	switch a.focusedPane {
	case RunsPane:
		a.focusedPane = WorkflowsPane
	case JobsPane:
		a.focusedPane = RunsPane
	}
}

// focusNextPane moves focus to the next pane
func (a *App) focusNextPane() {
	switch a.focusedPane {
	case WorkflowsPane:
		a.focusedPane = RunsPane
	case RunsPane:
		a.focusedPane = JobsPane
	}
}

// focusPrevPaneWithSelect moves to previous panel and triggers data loading
func (a *App) focusPrevPaneWithSelect() tea.Cmd {
	switch a.focusedPane {
	case RunsPane:
		a.focusedPane = WorkflowsPane
		return a.onWorkflowSelectionChange()
	case JobsPane:
		a.focusedPane = RunsPane
		return a.onRunSelectionChange()
	}
	return nil
}

// focusNextPaneWithSelect moves to next panel and triggers data loading
func (a *App) focusNextPaneWithSelect() tea.Cmd {
	switch a.focusedPane {
	case WorkflowsPane:
		a.focusedPane = RunsPane
		return a.onRunSelectionChange()
	case RunsPane:
		a.focusedPane = JobsPane
		return a.onJobSelectionChange()
	}
	return nil
}
