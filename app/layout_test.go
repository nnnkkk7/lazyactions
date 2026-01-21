package app

import (
	"testing"
)

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

func TestApp_PanelLayout(t *testing.T) {
	app := New()
	app.height = 50

	totalHeight, panelHeight := app.panelLayout()
	// totalHeight = height - StatusBarHeight = 50 - 1 = 49
	if totalHeight != 49 {
		t.Errorf("panelLayout totalHeight = %d, want 49", totalHeight)
	}
	// panelHeight = totalHeight / NumLeftPanels = 49 / 3 = 16
	if panelHeight != 16 {
		t.Errorf("panelLayout panelHeight = %d, want 16", panelHeight)
	}

	logHeight := app.logPaneHeight()
	// logPaneHeight = height - StatusAreaHeight = 50 - 2 = 48
	if logHeight != 48 {
		t.Errorf("logPaneHeight = %d, want 48", logHeight)
	}
}
