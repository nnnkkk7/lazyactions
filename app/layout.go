package app

// Layout calculation functions for the TUI application.
// These functions calculate dimensions and positions for panels.

func (a *App) leftPanelWidth() int {
	w := int(float64(a.width) * LeftPanelWidthRatio)
	if w < MinLeftPanelWidth {
		w = MinLeftPanelWidth
	}
	return w
}

func (a *App) logPaneWidth() int {
	// 50% of width for logs pane
	w := int(float64(a.width) * LogPaneWidthRatio)
	if w < MinLogPaneWidth {
		w = MinLogPaneWidth
	}
	return w
}

func (a *App) logPaneHeight() int {
	return a.height - StatusAreaHeight
}

// panelLayout returns the total height and individual panel height for the left sidebar
func (a *App) panelLayout() (totalHeight, panelHeight int) {
	totalHeight = a.height - StatusBarHeight
	if totalHeight < MinTotalHeight {
		totalHeight = MinTotalHeight
	}
	panelHeight = totalHeight / NumLeftPanels
	if panelHeight < MinPanelHeight {
		panelHeight = MinPanelHeight
	}
	return totalHeight, panelHeight
}

// panelStartY returns the starting Y position for a given pane
func (a *App) panelStartY(pane Pane) int {
	_, panelHeight := a.panelLayout()
	switch pane {
	case WorkflowsPane:
		return 0
	case RunsPane:
		return panelHeight
	case JobsPane:
		return 2 * panelHeight
	default:
		return 0
	}
}

func (a *App) workflowsPaneWidth() int {
	// 20% of width for workflows pane
	w := int(float64(a.width) * WorkflowsPaneWidthRatio)
	if w < MinWorkflowsPaneWidth {
		w = MinWorkflowsPaneWidth
	}
	return w
}

func (a *App) runsPaneWidth() int {
	// Remaining width for runs pane
	w := a.width - a.workflowsPaneWidth() - a.logPaneWidth()
	if w < MinRunsPaneWidth {
		w = MinRunsPaneWidth
	}
	return w
}
