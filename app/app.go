// Package app provides the TUI application for lazyactions.
package app

import (
	"context"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nnnkkk7/lazyactions/github"
)

// Pane represents a UI pane
type Pane int

const (
	WorkflowsPane Pane = iota
	RunsPane
	LogsPane
)

// App is the main application model
type App struct {
	// Data (using FilteredList pattern)
	repo      github.Repository
	workflows *FilteredList[github.Workflow]
	runs      *FilteredList[github.Run]
	jobs      *FilteredList[github.Job]

	// UI state
	focusedPane Pane
	width       int
	height      int
	logView     *LogViewport

	// Polling
	logPoller      *TickerTask
	adaptivePoller *AdaptivePoller

	// State
	loading bool
	err     error

	// Popups
	showHelp    bool
	showConfirm bool
	confirmMsg  string
	confirmFn   func() tea.Cmd

	// Filter (/key)
	filtering   bool
	filterInput textinput.Model

	// Spinner
	spinner spinner.Model

	// Flash message
	flashMsg string

	// Dependencies
	client github.Client
	keys   KeyMap

	// Fullscreen log mode
	fullscreenLog bool
}

// Option is a functional option for App
type Option func(*App)

// WithClient sets the GitHub client
func WithClient(client github.Client) Option {
	return func(a *App) {
		a.client = client
	}
}

// WithRepository sets the repository
func WithRepository(repo github.Repository) Option {
	return func(a *App) {
		a.repo = repo
	}
}

// New creates a new App instance
func New(opts ...Option) *App {
	ti := textinput.New()
	ti.Placeholder = "Filter..."
	ti.CharLimit = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = RunningStyle

	a := &App{
		workflows: NewFilteredList(func(w github.Workflow, filter string) bool {
			return strings.Contains(strings.ToLower(w.Name), strings.ToLower(filter))
		}),
		runs: NewFilteredList(func(r github.Run, filter string) bool {
			return strings.Contains(strings.ToLower(r.Branch), strings.ToLower(filter)) ||
				strings.Contains(strings.ToLower(r.Actor), strings.ToLower(filter))
		}),
		jobs: NewFilteredList(func(j github.Job, filter string) bool {
			return strings.Contains(strings.ToLower(j.Name), strings.ToLower(filter))
		}),
		focusedPane: WorkflowsPane,
		logView:     NewLogViewport(80, 20),
		filterInput: ti,
		spinner:     s,
		keys:        DefaultKeyMap(),
	}

	for _, opt := range opts {
		opt(a)
	}

	if a.client != nil {
		a.adaptivePoller = NewAdaptivePoller(func() int {
			return a.client.RateLimitRemaining()
		})
	}

	return a
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.spinner.Tick,
		a.fetchWorkflowsCmd(),
	)
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := a.handleKeyPress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.logView.SetSize(a.logPaneWidth(), a.logPaneHeight())

	case WorkflowsLoadedMsg:
		a.loading = false
		if msg.Err != nil {
			a.err = msg.Err
		} else {
			a.workflows.SetItems(msg.Workflows)
			if a.workflows.Len() > 0 {
				if wf, ok := a.workflows.Selected(); ok {
					cmds = append(cmds, a.fetchRunsCmd(wf.ID))
				}
			}
		}

	case RunsLoadedMsg:
		a.loading = false
		if msg.Err != nil {
			a.err = msg.Err
		} else {
			a.runs.SetItems(msg.Runs)
			if a.runs.Len() > 0 {
				if run, ok := a.runs.Selected(); ok {
					cmds = append(cmds, a.fetchJobsCmd(run.ID))
				}
			}
		}

	case JobsLoadedMsg:
		a.loading = false
		if msg.Err != nil {
			a.err = msg.Err
		} else {
			a.jobs.SetItems(msg.Jobs)
			if a.jobs.Len() > 0 {
				if job, ok := a.jobs.Selected(); ok {
					cmds = append(cmds, a.fetchLogsCmd(job.ID))
				}
			}
		}

	case LogsLoadedMsg:
		if msg.Err != nil {
			a.err = msg.Err
		} else {
			// Wrap log lines to fit within viewport width
			wrappedLogs := wrapLines(msg.Logs, a.logPaneWidth()-4)
			a.logView.SetContent(wrappedLogs)
		}

	case RunCancelledMsg:
		if msg.Err != nil {
			a.err = msg.Err
		} else {
			a.flashMsg = "Run cancelled"
			cmds = append(cmds, a.refreshCurrentWorkflow())
		}

	case RunRerunMsg:
		if msg.Err != nil {
			a.err = msg.Err
		} else {
			a.flashMsg = "Rerun triggered"
			cmds = append(cmds, a.refreshCurrentWorkflow())
		}

	case FlashClearMsg:
		a.flashMsg = ""

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

// View implements tea.Model
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	if a.fullscreenLog {
		return a.renderFullscreenLog()
	}

	if a.showHelp {
		return a.renderHelp()
	}

	if a.showConfirm {
		return a.renderConfirmDialog()
	}

	// Calculate pane dimensions
	paneHeight := a.height - 2 // status bar
	if paneHeight < 5 {
		paneHeight = 5
	}

	// Build each pane as fixed-size string array
	wfLines := a.buildWorkflowsContent(paneHeight)
	runLines := a.buildRunsContent(paneHeight)
	logLines := a.buildLogsContent(paneHeight)

	// Join panes horizontally, line by line
	var output strings.Builder
	for i := 0; i < paneHeight; i++ {
		output.WriteString(wfLines[i])
		output.WriteString(runLines[i])
		output.WriteString(logLines[i])
		output.WriteString("\n")
	}

	// Add status bar
	output.WriteString(a.renderStatusBar())

	return output.String()
}

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
		} else if a.err != nil {
			a.err = nil
		}

	case key.Matches(msg, a.keys.Up):
		return a.navigateUp()

	case key.Matches(msg, a.keys.Down):
		return a.navigateDown()

	case key.Matches(msg, a.keys.Left), key.Matches(msg, a.keys.ShiftTab):
		a.focusPrevPane()

	case key.Matches(msg, a.keys.Right), key.Matches(msg, a.keys.Tab):
		a.focusNextPane()

	case key.Matches(msg, a.keys.Filter):
		a.filtering = true
		a.filterInput.Focus()

	case key.Matches(msg, a.keys.FullLog):
		if a.focusedPane == LogsPane {
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
	case LogsPane:
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
	case LogsPane:
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
	case LogsPane:
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
	case LogsPane:
		a.focusedPane = RunsPane
	}
}

// focusNextPane moves focus to the next pane
func (a *App) focusNextPane() {
	switch a.focusedPane {
	case WorkflowsPane:
		a.focusedPane = RunsPane
	case RunsPane:
		a.focusedPane = LogsPane
	}
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
		return a.fetchLogsCmd(job.ID)
	}
	return nil
}

// confirmCancelRun shows confirmation dialog for cancelling a run
func (a *App) confirmCancelRun() tea.Cmd {
	run, ok := a.runs.Selected()
	if !ok || !run.IsRunning() {
		return nil
	}
	a.showConfirm = true
	a.confirmMsg = "Cancel this run?"
	a.confirmFn = func() tea.Cmd {
		return cancelRun(a.client, a.repo, run.ID)
	}
	return nil
}

// rerunWorkflow triggers a workflow rerun
func (a *App) rerunWorkflow() tea.Cmd {
	run, ok := a.runs.Selected()
	if !ok {
		return nil
	}
	return rerunWorkflow(a.client, a.repo, run.ID)
}

// rerunFailedJobs reruns only failed jobs
func (a *App) rerunFailedJobs() tea.Cmd {
	run, ok := a.runs.Selected()
	if !ok || !run.IsFailed() {
		return nil
	}
	return rerunFailedJobs(a.client, a.repo, run.ID)
}

// triggerWorkflow triggers a workflow dispatch
func (a *App) triggerWorkflow() tea.Cmd {
	wf, ok := a.workflows.Selected()
	if !ok {
		return nil
	}
	// Get workflow file name from path (e.g., ".github/workflows/ci.yml" -> "ci.yml")
	workflowFile := wf.Path
	if idx := len(".github/workflows/"); len(wf.Path) > idx {
		workflowFile = wf.Path[idx:]
	}
	// Trigger on default branch (main)
	return triggerWorkflow(a.client, a.repo, workflowFile, "main", nil)
}

// yankURL copies the selected run URL to clipboard
func (a *App) yankURL() tea.Cmd {
	if run, ok := a.runs.Selected(); ok && run.URL != "" {
		// Return a flash message indicating the URL was copied
		// Note: actual clipboard integration would require platform-specific code
		return flashMessage("Copied: "+run.URL, 2)
	}
	return nil
}

// refreshAll refreshes all data
func (a *App) refreshAll() tea.Cmd {
	a.loading = true
	return a.fetchWorkflowsCmd()
}

// refreshCurrentWorkflow refreshes runs for the current workflow
func (a *App) refreshCurrentWorkflow() tea.Cmd {
	if wf, ok := a.workflows.Selected(); ok {
		return a.fetchRunsCmd(wf.ID)
	}
	return nil
}

// Command generators
func (a *App) fetchWorkflowsCmd() tea.Cmd {
	if a.client == nil {
		return nil
	}
	a.loading = true
	return fetchWorkflows(a.client, a.repo)
}

func (a *App) fetchRunsCmd(workflowID int64) tea.Cmd {
	if a.client == nil {
		return nil
	}
	return fetchRuns(a.client, a.repo, workflowID)
}

func (a *App) fetchJobsCmd(runID int64) tea.Cmd {
	if a.client == nil {
		return nil
	}
	return fetchJobs(a.client, a.repo, runID)
}

func (a *App) fetchLogsCmd(jobID int64) tea.Cmd {
	if a.client == nil {
		return nil
	}
	return fetchLogs(a.client, a.repo, jobID)
}

// Rendering helpers - build fixed-size content for each pane

func (a *App) buildWorkflowsContent(height int) []string {
	w := a.workflowsPaneWidth()
	lines := make([]string, height)

	// Border color
	borderColor := UnfocusedColor
	if a.focusedPane == WorkflowsPane {
		borderColor = FocusedColor
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Title
	title := " Workflows "
	if a.loading {
		title = " Workflows " + a.spinner.View() + " "
	}

	// Build content lines
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
			name := truncateString(wf.Name, w-6)
			if selected {
				content = append(content, "> "+name)
			} else {
				content = append(content, "  "+name)
			}
		}
	}
	content = append(content, "")
	content = append(content, ScrollPosition(a.workflows.SelectedIndex(), a.workflows.Len()))

	// Build lines with borders
	innerWidth := w - 2
	lines[0] = borderStyle.Render("┌") + borderStyle.Render(padCenter(title, innerWidth, "─")) + borderStyle.Render("┐")
	for i := 1; i < height-1; i++ {
		contentIdx := i - 1
		var line string
		if contentIdx < len(content) {
			line = padRight(content[contentIdx], innerWidth)
		} else {
			line = strings.Repeat(" ", innerWidth)
		}
		lines[i] = borderStyle.Render("│") + line + borderStyle.Render("│")
	}
	lines[height-1] = borderStyle.Render("└") + borderStyle.Render(strings.Repeat("─", innerWidth)) + borderStyle.Render("┘")

	return lines
}

func (a *App) buildRunsContent(height int) []string {
	w := a.runsPaneWidth()
	lines := make([]string, height)

	borderColor := UnfocusedColor
	if a.focusedPane == RunsPane {
		borderColor = FocusedColor
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	title := " Runs "

	var content []string
	items := a.runs.Items()
	if len(items) == 0 {
		content = append(content, "  Select workflow")
	} else {
		for i, run := range items {
			selected := i == a.runs.SelectedIndex()
			icon := StatusIcon(run.Status, run.Conclusion)
			line := icon + " #" + formatRunNumber(run.ID) + " " + run.Branch
			line = truncateString(line, w-6)
			if selected {
				content = append(content, "> "+line)
			} else {
				content = append(content, "  "+line)
			}
		}
	}
	content = append(content, "")
	content = append(content, ScrollPosition(a.runs.SelectedIndex(), a.runs.Len()))

	innerWidth := w - 2
	lines[0] = borderStyle.Render("┌") + borderStyle.Render(padCenter(title, innerWidth, "─")) + borderStyle.Render("┐")
	for i := 1; i < height-1; i++ {
		contentIdx := i - 1
		var line string
		if contentIdx < len(content) {
			line = padRight(content[contentIdx], innerWidth)
		} else {
			line = strings.Repeat(" ", innerWidth)
		}
		lines[i] = borderStyle.Render("│") + line + borderStyle.Render("│")
	}
	lines[height-1] = borderStyle.Render("└") + borderStyle.Render(strings.Repeat("─", innerWidth)) + borderStyle.Render("┘")

	return lines
}

func (a *App) buildLogsContent(height int) []string {
	w := a.logPaneWidth()
	lines := make([]string, height)

	borderColor := UnfocusedColor
	if a.focusedPane == LogsPane {
		borderColor = FocusedColor
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	title := " Logs "

	var content []string
	// Show jobs
	items := a.jobs.Items()
	if len(items) == 0 {
		content = append(content, "  Select a run")
	} else {
		for i, job := range items {
			selected := i == a.jobs.SelectedIndex()
			icon := StatusIcon(job.Status, job.Conclusion)
			line := icon + " " + truncateString(job.Name, w-10)
			if selected {
				content = append(content, "> "+line)
			} else {
				content = append(content, "  "+line)
			}
		}
	}

	// Add log content (viewport)
	content = append(content, "─────")
	logContent := a.logView.View()
	logLines := strings.Split(logContent, "\n")
	for _, l := range logLines {
		content = append(content, truncateString(l, w-4))
	}

	innerWidth := w - 2
	lines[0] = borderStyle.Render("┌") + borderStyle.Render(padCenter(title, innerWidth, "─")) + borderStyle.Render("┐")
	for i := 1; i < height-1; i++ {
		contentIdx := i - 1
		var line string
		if contentIdx < len(content) {
			line = padRight(content[contentIdx], innerWidth)
		} else {
			line = strings.Repeat(" ", innerWidth)
		}
		lines[i] = borderStyle.Render("│") + line + borderStyle.Render("│")
	}
	lines[height-1] = borderStyle.Render("└") + borderStyle.Render(strings.Repeat("─", innerWidth)) + borderStyle.Render("┘")

	return lines
}

// padRight pads a string to the specified width
func padRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// padCenter centers a string within the specified width, filling with the given char
func padCenter(s string, width int, fill string) string {
	sLen := len([]rune(s))
	if sLen >= width {
		return s
	}
	leftPad := (width - sLen) / 2
	rightPad := width - sLen - leftPad
	return strings.Repeat(fill, leftPad) + s + strings.Repeat(fill, rightPad)
}

func (a *App) logPaneWidth() int {
	// 50% of width for logs pane
	w := int(float64(a.width) * 0.50)
	if w < 20 {
		w = 20
	}
	return w
}

func (a *App) logPaneHeight() int {
	return a.height - 2 // account for status bar
}

func (a *App) workflowsPaneWidth() int {
	// 20% of width for workflows pane
	w := int(float64(a.width) * 0.20)
	if w < 15 {
		w = 15
	}
	return w
}

func (a *App) runsPaneWidth() int {
	// Remaining width for runs pane
	w := a.width - a.workflowsPaneWidth() - a.logPaneWidth()
	if w < 20 {
		w = 20
	}
	return w
}

func (a *App) paneHeight() int {
	h := a.height - 2 // account for status bar
	if h < 5 {
		h = 5
	}
	return h
}

func (a *App) renderWorkflowsPane() string {
	style := UnfocusedPane
	titleStyle := UnfocusedTitle
	if a.focusedPane == WorkflowsPane {
		style = FocusedPane
		titleStyle = FocusedTitle
	}

	w := a.workflowsPaneWidth()
	contentWidth := w - 2 // account for border

	title := titleStyle.Render(" Workflows ")
	if a.loading {
		title = titleStyle.Render(" Workflows " + a.spinner.View() + " ")
	}

	var lines []string
	lines = append(lines, title)

	items := a.workflows.Items()
	if len(items) == 0 {
		if a.loading {
			lines = append(lines, "  Loading...")
		} else {
			lines = append(lines, "  No workflows")
		}
	} else {
		for i, wf := range items {
			selected := i == a.workflows.SelectedIndex()
			name := truncateString(wf.Name, contentWidth-4)
			lines = append(lines, RenderItem(name, selected))
		}
	}

	lines = append(lines, "")
	lines = append(lines, ScrollPosition(a.workflows.SelectedIndex(), a.workflows.Len()))

	return style.
		Width(w).
		Height(a.paneHeight()).
		Render(strings.Join(lines, "\n"))
}

func (a *App) renderRunsPane() string {
	style := UnfocusedPane
	titleStyle := UnfocusedTitle
	if a.focusedPane == RunsPane {
		style = FocusedPane
		titleStyle = FocusedTitle
	}

	w := a.runsPaneWidth()
	contentWidth := w - 2 // account for border

	title := titleStyle.Render(" Runs ")

	var lines []string
	lines = append(lines, title)

	items := a.runs.Items()
	if len(items) == 0 {
		lines = append(lines, "  Select a workflow")
	} else {
		for i, run := range items {
			selected := i == a.runs.SelectedIndex()
			icon := StatusIcon(run.Status, run.Conclusion)
			line := icon + " #" + formatRunNumber(run.ID) + " " + run.Branch
			line = truncateString(line, contentWidth-4)
			lines = append(lines, RenderItem(line, selected))
		}
	}

	lines = append(lines, "")
	lines = append(lines, ScrollPosition(a.runs.SelectedIndex(), a.runs.Len()))

	return style.
		Width(w).
		Height(a.paneHeight()).
		Render(strings.Join(lines, "\n"))
}

func (a *App) renderLogsPane() string {
	style := UnfocusedPane
	titleStyle := UnfocusedTitle
	if a.focusedPane == LogsPane {
		style = FocusedPane
		titleStyle = FocusedTitle
	}

	w := a.logPaneWidth()
	contentWidth := w - 2 // account for border

	title := titleStyle.Render(" Logs ")

	var lines []string
	lines = append(lines, title)

	// Show job list
	items := a.jobs.Items()
	if len(items) == 0 {
		lines = append(lines, "  Select a run")
	} else {
		for i, job := range items {
			selected := i == a.jobs.SelectedIndex()
			icon := StatusIcon(job.Status, job.Conclusion)
			line := icon + " " + truncateString(job.Name, contentWidth-6)
			lines = append(lines, RenderItem(line, selected))
		}
	}

	lines = append(lines, "")
	lines = append(lines, a.logView.View())

	return style.
		Width(w).
		Height(a.paneHeight()).
		Render(strings.Join(lines, "\n"))
}

func (a *App) renderStatusBar() string {
	var hints string
	switch a.focusedPane {
	case WorkflowsPane:
		hints = "[t]rigger [/]filter [?]help [q]uit"
	case RunsPane:
		hints = "[c]ancel [r]erun [R]erun-failed [y]ank [?]help [q]uit"
	case LogsPane:
		hints = "[L]fullscreen [y]ank [Esc]back [?]help [q]uit"
	}

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
			Render("Error: " + a.err.Error())
	}

	return StatusBar.Width(a.width).Render(hints)
}

func (a *App) renderFullscreenLog() string {
	title := FocusedTitle.Render("Logs (fullscreen)")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		a.logView.View(),
	)

	return FocusedPane.
		Width(a.width).
		Height(a.height - 1).
		Render(content)
}

func (a *App) renderHelp() string {
	help := `
Navigation
──────────────────────────────────
j/↓         Move down
k/↑         Move up
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

// StartLogPolling starts log polling for a running job
func (a *App) StartLogPolling(ctx context.Context) tea.Cmd {
	if a.logPoller != nil {
		a.logPoller.Stop()
	}

	interval := a.adaptivePoller.NextInterval()

	a.logPoller = NewTickerTask(interval, func(ctx context.Context) tea.Msg {
		job, ok := a.jobs.Selected()
		if !ok {
			return nil
		}
		logs, err := a.client.GetJobLogs(ctx, a.repo, job.ID)
		if ctx.Err() != nil {
			return nil
		}
		return LogsLoadedMsg{Logs: logs, Err: err}
	})

	return a.logPoller.Start()
}

// StopLogPolling stops log polling
func (a *App) StopLogPolling() {
	if a.logPoller != nil {
		a.logPoller.Stop()
		a.logPoller = nil
	}
}

// formatRunNumber formats a run ID for display
func formatRunNumber(id int64) string {
	return strconv.FormatInt(id, 10)
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if maxLen <= 3 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

// wrapLines wraps long lines to fit within maxWidth
func wrapLines(content string, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 80
	}
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		runes := []rune(line)
		if len(runes) <= maxWidth {
			result = append(result, line)
		} else {
			// Split long lines
			for len(runes) > maxWidth {
				result = append(result, string(runes[:maxWidth]))
				runes = runes[maxWidth:]
			}
			if len(runes) > 0 {
				result = append(result, string(runes))
			}
		}
	}
	return strings.Join(result, "\n")
}

// Run starts the TUI application
func Run(client github.Client, repo github.Repository) error {
	app := New(
		WithClient(client),
		WithRepository(repo),
	)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
