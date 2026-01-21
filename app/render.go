package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Rendering helpers - build panels for lazygit-style layout

// buildWorkflowsPanel builds the workflows panel for the left sidebar
func (a *App) buildWorkflowsPanel(width, height int) []string {
	focused := a.focusedPane == WorkflowsPane
	borderStyle := getPanelBorderStyle(focused)

	// Title with spinner when loading
	titleText := "Workflows"
	if a.loading {
		titleText = "Workflows " + a.spinner.View()
	}
	title := renderPanelTitle(titleText, focused)

	// Build content
	leftWidth := a.leftPanelWidth()
	var content []string
	items := a.workflows.Items()
	if len(items) == 0 {
		if a.loading {
			content = append(content, "  Loading...")
		} else {
			content = append(content, "  No workflows")
		}
	} else {
		for i, wf := range items {
			selected := i == a.workflows.SelectedIndex()
			hovered := a.mouseX < leftWidth && a.mouseY == i+BorderOffset
			name := truncateString(wf.Name, width-ItemPaddingSmall)
			content = append(content, a.renderListItem(name, selected, focused, hovered))
		}
	}

	return renderPanelFrame(width, height, title, content, borderStyle)
}

// buildRunsPanel builds the runs panel for the left sidebar
func (a *App) buildRunsPanel(width, height int) []string {
	focused := a.focusedPane == RunsPane
	borderStyle := getPanelBorderStyle(focused)
	title := renderPanelTitle("Runs", focused)

	// Calculate panel position for hover detection
	leftWidth := a.leftPanelWidth()
	panelStartY := a.panelStartY(RunsPane)

	// Build content
	var content []string
	items := a.runs.Items()
	if len(items) == 0 {
		content = append(content, "  Select workflow")
	} else {
		for i, run := range items {
			selected := i == a.runs.SelectedIndex()
			hovered := a.mouseX < leftWidth && a.mouseY == panelStartY+i+BorderOffset
			icon := StatusIcon(run.Status, run.Conclusion)
			line := icon + " #" + formatRunNumber(run.ID) + " " + run.Branch
			line = truncateString(line, width-ItemPaddingSmall)
			content = append(content, a.renderListItem(line, selected, focused, hovered))
		}
	}

	return renderPanelFrame(width, height, title, content, borderStyle)
}

// buildJobsPanel builds the jobs panel for the left sidebar
func (a *App) buildJobsPanel(width, height int) []string {
	focused := a.focusedPane == JobsPane
	borderStyle := getPanelBorderStyle(focused)
	title := renderPanelTitle("Jobs", focused)

	// Calculate panel position for hover detection
	leftWidth := a.leftPanelWidth()
	panelStartY := a.panelStartY(JobsPane)

	// Build content
	var content []string
	items := a.jobs.Items()
	if len(items) == 0 {
		content = append(content, "  Select a run")
	} else {
		for i, job := range items {
			selected := i == a.jobs.SelectedIndex()
			hovered := a.mouseX < leftWidth && a.mouseY == panelStartY+i+BorderOffset
			icon := StatusIcon(job.Status, job.Conclusion)
			line := icon + " " + truncateString(job.Name, width-ItemPaddingMedium)
			content = append(content, a.renderListItem(line, selected, focused, hovered))
		}
	}

	return renderPanelFrame(width, height, title, content, borderStyle)
}

// buildDetailPanel builds the detail view panel (right side) with tabs
func (a *App) buildDetailPanel(width, height int) []string {
	borderStyle := getPanelBorderStyle(false) // Detail panel is always unfocused style

	// Build tab header
	infoTab := " Info "
	logsTab := " Logs "
	if a.detailTab == InfoTab {
		infoTab = FocusedTitle.Render(" Info ")
	} else {
		logsTab = FocusedTitle.Render(" Logs ")
	}
	tabHeader := " [1]" + infoTab + " [2]" + logsTab + " "

	// Build content based on selected tab
	var content []string
	if a.detailTab == InfoTab {
		content = a.buildInfoContent(width - ContentPadding)
	} else {
		content = a.buildLogsContent(width - ContentPadding)
	}

	return renderPanelFrame(width, height, tabHeader, content, borderStyle)
}

// buildInfoContent builds the content for the Info tab
func (a *App) buildInfoContent(maxWidth int) []string {
	var content []string

	switch a.focusedPane {
	case WorkflowsPane:
		if wf, ok := a.workflows.Selected(); ok {
			content = append(content, "  Workflow Information")
			content = append(content, "  "+strings.Repeat("─", 30))
			content = append(content, "  Name:  "+wf.Name)
			content = append(content, "  Path:  "+wf.Path)
			content = append(content, "  State: "+wf.State)
		} else {
			content = append(content, "  Select a workflow")
		}

	case RunsPane:
		if run, ok := a.runs.Selected(); ok {
			content = append(content, "  Run Information")
			content = append(content, "  "+strings.Repeat("─", 30))
			content = append(content, "  ID:     #"+formatRunNumber(run.ID))
			content = append(content, "  Status: "+StatusIcon(run.Status, run.Conclusion)+" "+run.Status)
			if run.Conclusion != "" {
				content = append(content, "  Result: "+run.Conclusion)
			}
			content = append(content, "  Branch: "+run.Branch)
			content = append(content, "  Event:  "+run.Event)
			content = append(content, "  Actor:  "+run.Actor)
			if !run.CreatedAt.IsZero() {
				content = append(content, "  Created: "+run.CreatedAt.Format("2006-01-02 15:04:05"))
			}
			if run.URL != "" {
				content = append(content, "")
				content = append(content, "  URL: "+truncateString(run.URL, maxWidth-6))
			}
		} else {
			content = append(content, "  Select a run")
		}

	case JobsPane:
		if job, ok := a.jobs.Selected(); ok {
			content = append(content, "  Job Information")
			content = append(content, "  "+strings.Repeat("─", 30))
			content = append(content, "  Name:   "+job.Name)
			content = append(content, "  Status: "+StatusIcon(job.Status, job.Conclusion)+" "+job.Status)
			if job.Conclusion != "" {
				content = append(content, "  Result: "+job.Conclusion)
			}
			if len(job.Steps) > 0 {
				content = append(content, "")
				content = append(content, "  Steps:")
				for _, step := range job.Steps {
					icon := StatusIcon(step.Status, step.Conclusion)
					content = append(content, "    "+icon+" "+truncateString(step.Name, maxWidth-10))
				}
			}
		} else {
			content = append(content, "  Select a job")
		}
	}

	return content
}

// buildLogsContent builds the content for the Logs tab
func (a *App) buildLogsContent(maxWidth int) []string {
	var content []string

	job, jobOk := a.jobs.Selected()
	if jobOk {
		content = append(content, "  Logs: "+job.Name)
		content = append(content, "  "+strings.Repeat("─", 30))
	}

	// Show step selection list if we have parsed steps
	if a.parsedLogs != nil && len(a.parsedLogs.Steps) > 0 {
		// Navigation hint
		if a.stepListFocused {
			content = append(content, "  Steps: (↑/↓ select, Enter focus logs)")
		} else {
			content = append(content, "  Steps: (Esc back to steps)")
		}
		content = append(content, "")

		// "All logs" option
		allLogsSelected := a.selectedStepIdx == -1
		allLogsText := "All logs"
		if allLogsSelected {
			if a.stepListFocused {
				content = append(content, "  "+CursorStyle.Render(">")+" "+SelectedItemFocused.Render(allLogsText))
			} else {
				content = append(content, "  "+SelectedItemUnfocused.Render("> "+allLogsText))
			}
		} else {
			content = append(content, "    "+NormalItem.Render(allLogsText))
		}

		// Step list with status icons
		for i, step := range a.parsedLogs.Steps {
			stepSelected := a.selectedStepIdx == i

			// Get step status from job.Steps if available
			icon := " "
			if jobOk && i < len(job.Steps) {
				icon = StatusIcon(job.Steps[i].Status, job.Steps[i].Conclusion)
			}

			stepName := truncateString(step.Name, maxWidth-10)
			stepText := icon + " " + stepName

			if stepSelected {
				if a.stepListFocused {
					content = append(content, "  "+CursorStyle.Render(">")+" "+SelectedItemFocused.Render(stepText))
				} else {
					content = append(content, "  "+SelectedItemUnfocused.Render("> "+stepText))
				}
			} else {
				content = append(content, "    "+NormalItem.Render(stepText))
			}
		}

		content = append(content, "")
		content = append(content, "  "+strings.Repeat("─", 30))
	}

	// Log content
	logContent := a.logView.View()
	logLines := strings.Split(logContent, "\n")
	for _, l := range logLines {
		content = append(content, "  "+truncateString(l, maxWidth-4))
	}

	if len(content) == 0 {
		content = append(content, "  No logs available")
	}

	return content
}

// padRight pads a string to the specified display width
func padRight(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		// Truncate if too long
		return truncateToWidth(s, width)
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

// getPanelBorderStyle returns the border style based on focus state
func getPanelBorderStyle(focused bool) lipgloss.Style {
	borderColor := UnfocusedColor
	if focused {
		borderColor = FocusedColor
	}
	return lipgloss.NewStyle().Foreground(borderColor)
}

// renderPanelTitle renders the title with lazydocker-style inverted colors when focused
func renderPanelTitle(titleText string, focused bool) string {
	if focused {
		return " " + FocusedTitle.Render(" "+titleText+" ") + " "
	}
	return " " + titleText + " "
}

// renderPanelFrame renders a panel frame with the given content
// This is a common pattern used by all panel builders
func renderPanelFrame(width, height int, title string, content []string, borderStyle lipgloss.Style) []string {
	if height < BorderWidth {
		return []string{}
	}
	lines := make([]string, height)
	innerWidth := width - BorderWidth

	lines[0] = buildBorderHeader(title, innerWidth, borderStyle)
	for i := 1; i < height-1; i++ {
		contentIdx := i - 1
		var line string
		if contentIdx < len(content) {
			line = padRight(content[contentIdx], innerWidth)
		} else {
			line = strings.Repeat(" ", innerWidth)
		}
		lines[i] = borderStyle.Render("┃") + line + borderStyle.Render("┃")
	}
	lines[height-1] = borderStyle.Render("┗" + strings.Repeat("━", innerWidth) + "┛")

	return lines
}

// buildBorderHeader builds a header line with proper styling for borders
// This ensures the border color is applied correctly even after styled title text
func buildBorderHeader(title string, innerWidth int, borderStyle lipgloss.Style) string {
	titleWidth := lipgloss.Width(title)
	if titleWidth >= innerWidth {
		return borderStyle.Render("┏") + title + borderStyle.Render("┓")
	}
	leftPad := (innerWidth - titleWidth) / 2
	rightPad := innerWidth - titleWidth - leftPad
	return borderStyle.Render("┏"+strings.Repeat("━", leftPad)) + title + borderStyle.Render(strings.Repeat("━", rightPad)+"┓")
}

// truncateToWidth truncates a string to fit within the specified display width
func truncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	currentWidth := 0
	for i, r := range s {
		charWidth := lipgloss.Width(string(r))
		if currentWidth+charWidth > maxWidth {
			if maxWidth >= 3 && currentWidth >= 0 {
				return s[:i] + "..."
			}
			return s[:i]
		}
		currentWidth += charWidth
	}
	return s
}

// renderStatusBar renders the status bar at the bottom
func (a *App) renderStatusBar() string {
	// Navigation hints
	navHints := "[j/k]panel [↑/↓]list"

	// Pane-specific action hints
	var actionHints string
	switch a.focusedPane {
	case WorkflowsPane:
		actionHints = "[t]rigger [/]filter"
	case RunsPane:
		actionHints = "[c]ancel [r]erun [R]erun-failed [y]ank"
	case JobsPane:
		if a.detailTab == LogsTab && a.parsedLogs != nil && len(a.parsedLogs.Steps) > 0 {
			if a.stepListFocused {
				actionHints = "[↑/↓]step [Enter]logs [L]fullscreen"
			} else {
				actionHints = "[↑/↓]scroll [Esc]steps [L]fullscreen"
			}
		} else {
			actionHints = "[L]fullscreen [y]ank"
		}
	}

	// Tab hints
	tabHints := "[1]info [2]logs"

	// Common hints
	commonHints := "[?]help [q]uit"

	hints := navHints + " " + actionHints + " " + tabHints + " " + commonHints

	if a.filtering {
		return StatusBar.Width(a.width).Render("Filter: " + a.filterInput.View())
	}

	if a.flashMsg != "" {
		return StatusBar.Width(a.width).Render(a.flashMsg)
	}

	if a.err != nil {
		return StatusBar.
			Foreground(lipgloss.Color("#FF0000")).
			Width(a.width).
			Render("Error: " + a.err.Error() + " [Esc]retry")
	}

	return StatusBar.Width(a.width).Render(hints)
}

// renderFullscreenLog renders the fullscreen log view
func (a *App) renderFullscreenLog() string {
	title := FocusedTitle.Render("Logs (fullscreen)")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		a.logView.View(),
	)

	return FocusedPane.
		Width(a.width).
		Height(a.height - StatusBarHeight).
		Render(content)
}

// renderHelp renders the help popup
func (a *App) renderHelp() string {
	help := `
Panel Navigation
──────────────────────────────────
j           Next panel (down)
k           Previous panel (up)
↓/↑         Move in list (down/up)
h/←         Previous pane
l/→         Next pane
Tab         Next pane
Shift+Tab   Previous pane

Actions
──────────────────────────────────
t           Trigger workflow
c           Cancel run
r           Rerun workflow
R           Rerun failed jobs only
y           Copy URL to clipboard

Detail View
──────────────────────────────────
1           Info tab
2           Logs tab

Step Navigation (Logs tab)
──────────────────────────────────
↓/↑         Select step
Enter       Focus log content
Esc         Back to step list

View
──────────────────────────────────
/           Filter
L           Full-screen log
Esc         Close/Back
?           Toggle help
q           Quit
`
	return lipgloss.Place(a.width, a.height,
		lipgloss.Center, lipgloss.Center,
		HelpPopup.Render(help))
}

// renderConfirmDialog renders the confirmation dialog
func (a *App) renderConfirmDialog() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render(a.confirmMsg),
		"",
		"[y] Yes  [n] No",
	)
	dialog := ConfirmDialog.Width(40).Render(content)
	return lipgloss.Place(a.width, a.height,
		lipgloss.Center, lipgloss.Center, dialog)
}

// renderListItem renders a list item with appropriate styling based on selection and focus state
func (a *App) renderListItem(text string, selected, focused, _ bool) string {
	if selected {
		if focused {
			// Focused + selected: green cursor + bright selection
			return CursorStyle.Render(">") + SelectedItemFocused.Render(" "+text)
		}
		// Unfocused + selected: dim selection without cursor
		return SelectedItemUnfocused.Render("  " + text)
	}
	// Not selected: normal text
	return NormalItem.Render("  " + text)
}

// truncateString truncates a string to maxLen display width, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if maxLen <= 3 {
		return s
	}
	if lipgloss.Width(s) <= maxLen {
		return s
	}
	return truncateToWidth(s, maxLen)
}

// wrapLines wraps long lines to fit within maxWidth (display width)
func wrapLines(content string, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = DefaultWrapWidth
	}
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		if lipgloss.Width(line) <= maxWidth {
			result = append(result, line)
		} else {
			// Split long lines by display width
			currentLine := ""
			currentWidth := 0
			for _, r := range line {
				charWidth := lipgloss.Width(string(r))
				if currentWidth+charWidth > maxWidth {
					result = append(result, currentLine)
					currentLine = string(r)
					currentWidth = charWidth
				} else {
					currentLine += string(r)
					currentWidth += charWidth
				}
			}
			if currentLine != "" {
				result = append(result, currentLine)
			}
		}
	}
	return strings.Join(result, "\n")
}
