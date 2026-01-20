// Package app provides the TUI application for lazyactions.
package app

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
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
	JobsPane
)

// DetailTab represents the tab in the detail view
type DetailTab int

const (
	LogsTab DetailTab = iota
	InfoTab
)

// Clipboard is an interface for clipboard operations
type Clipboard interface {
	WriteAll(text string) error
}

// realClipboard implements Clipboard using the system clipboard
type realClipboard struct{}

func (c *realClipboard) WriteAll(text string) error {
	return clipboard.WriteAll(text)
}

// App is the main application model
type App struct {
	// Data (using FilteredList pattern)
	repo      github.Repository
	workflows *FilteredList[github.Workflow]
	runs      *FilteredList[github.Run]
	jobs      *FilteredList[github.Job]

	// UI state
	focusedPane Pane
	detailTab   DetailTab
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
	client    github.Client
	clipboard Clipboard
	keys      KeyMap

	// Fullscreen log mode
	fullscreenLog bool

	// Mouse tracking
	mouseX int
	mouseY int

	// Step-selectable logs
	parsedLogs      *ParsedLogs // Parsed log structure with steps
	selectedStepIdx int         // -1 = "All logs", 0+ = specific step
	stepListFocused bool        // Whether the step list has focus (vs log content)
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

// WithClipboard sets the clipboard implementation
func WithClipboard(cb Clipboard) Option {
	return func(a *App) {
		a.clipboard = cb
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
		focusedPane:     WorkflowsPane,
		logView:         NewLogViewport(80, 20),
		filterInput:     ti,
		spinner:         s,
		keys:            DefaultKeyMap(),
		selectedStepIdx: -1, // -1 means "All logs"
		stepListFocused: true,
	}

	for _, opt := range opts {
		opt(a)
	}

	// Set default clipboard if not provided
	if a.clipboard == nil {
		a.clipboard = &realClipboard{}
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

	case tea.MouseMsg:
		model, cmd := a.handleMouseEvent(msg)
		if cmd != nil {
			return model, cmd
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
		// Only update logs if they are for the currently selected job
		// This prevents stale logs from overwriting newer ones
		if job, ok := a.jobs.Selected(); ok && job.ID == msg.JobID {
			if msg.Err != nil {
				a.err = msg.Err
				a.logView.SetContent("Failed to load logs")
				a.parsedLogs = nil
			} else {
				// Parse logs into steps
				a.parsedLogs = ParseLogs(msg.Logs)
				// Update log view with selected step's logs (formatted)
				a.updateLogViewContent()
			}
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

	case RerunFailedJobsMsg:
		if msg.Err != nil {
			a.err = msg.Err
		} else {
			a.flashMsg = "Rerun failed jobs triggered"
			cmds = append(cmds, a.refreshCurrentWorkflow())
		}

	case WorkflowTriggeredMsg:
		if msg.Err != nil {
			a.err = msg.Err
		} else {
			a.flashMsg = "Workflow triggered: " + msg.Workflow
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

	// Calculate dimensions
	totalHeight := a.height - 1 // status bar
	if totalHeight < 10 {
		totalHeight = 10
	}

	// Left sidebar (30%), Right detail (70%)
	leftWidth := int(float64(a.width) * 0.30)
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := a.width - leftWidth

	// Left panels: 3 equal-height panels
	panelHeight := totalHeight / 3
	if panelHeight < 5 {
		panelHeight = 5
	}

	// Build left sidebar panels
	wfLines := a.buildWorkflowsPanel(leftWidth, panelHeight)
	runLines := a.buildRunsPanel(leftWidth, panelHeight)
	jobLines := a.buildJobsPanel(leftWidth, totalHeight-2*panelHeight) // remaining height

	// Build right detail view
	detailLines := a.buildDetailPanel(rightWidth, totalHeight)

	// Combine: left sidebar + right detail, line by line
	var output strings.Builder
	leftIdx := 0

	// Workflows panel
	for i := 0; i < panelHeight && leftIdx < totalHeight; i++ {
		line := wfLines[i]
		if leftIdx < len(detailLines) {
			line += detailLines[leftIdx]
		}
		output.WriteString(line)
		output.WriteString("\n")
		leftIdx++
	}

	// Runs panel
	for i := 0; i < panelHeight && leftIdx < totalHeight; i++ {
		line := runLines[i]
		if leftIdx < len(detailLines) {
			line += detailLines[leftIdx]
		}
		output.WriteString(line)
		output.WriteString("\n")
		leftIdx++
	}

	// Jobs panel (remaining height)
	jobHeight := totalHeight - 2*panelHeight
	for i := 0; i < jobHeight && leftIdx < totalHeight; i++ {
		line := jobLines[i]
		if leftIdx < len(detailLines) {
			line += detailLines[leftIdx]
		}
		output.WriteString(line)
		output.WriteString("\n")
		leftIdx++
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
	leftWidth := int(float64(a.width) * 0.30)
	if leftWidth < 20 {
		leftWidth = 20
	}
	totalHeight := a.height - 1
	if totalHeight < 10 {
		totalHeight = 10
	}
	panelHeight := totalHeight / 3
	if panelHeight < 5 {
		panelHeight = 5
	}

	// Handle clicks in the right panel (detail view)
	if x >= leftWidth {
		return a.handleDetailPanelClick(x, y, leftWidth, totalHeight)
	}

	// Determine which panel was clicked (left sidebar)
	if y < panelHeight {
		// Workflows panel
		a.focusedPane = WorkflowsPane
		// Calculate item index (account for border + title line)
		itemIdx := y - 2
		if itemIdx >= 0 && itemIdx < a.workflows.Len() {
			a.workflows.Select(itemIdx)
			return a, a.onWorkflowSelectionChange()
		}
	} else if y < 2*panelHeight {
		// Runs panel
		a.focusedPane = RunsPane
		itemIdx := y - panelHeight - 2
		if itemIdx >= 0 && itemIdx < a.runs.Len() {
			a.runs.Select(itemIdx)
			return a, a.onRunSelectionChange()
		}
	} else if y < totalHeight {
		// Jobs panel
		a.focusedPane = JobsPane
		itemIdx := y - 2*panelHeight - 2
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
	leftWidth := int(float64(a.width) * 0.30)
	if leftWidth < 20 {
		leftWidth = 20
	}

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
	leftWidth := int(float64(a.width) * 0.30)
	if leftWidth < 20 {
		leftWidth = 20
	}

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
	run, ok := a.runs.Selected()
	if !ok || run.URL == "" {
		return nil
	}

	if err := a.clipboard.WriteAll(run.URL); err != nil {
		// Clipboard not available (e.g., headless environment)
		// Show URL in flash message so user can copy manually
		return flashMessage("URL: "+run.URL, 3*time.Second)
	}
	return flashMessage("Copied: "+run.URL, 2*time.Second)
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

// Rendering helpers - build panels for lazygit-style layout

// buildWorkflowsPanel builds the workflows panel for the left sidebar
func (a *App) buildWorkflowsPanel(width, height int) []string {
	if height < 2 {
		return []string{}
	}
	lines := make([]string, height)

	focused := a.focusedPane == WorkflowsPane
	borderColor := UnfocusedColor
	if focused {
		borderColor = FocusedColor
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Title with lazydocker-style inverted colors when focused
	titleText := "Workflows"
	if a.loading {
		titleText = "Workflows " + a.spinner.View()
	}
	var title string
	if focused {
		title = " " + FocusedTitle.Render(" "+titleText+" ") + " "
	} else {
		title = " " + titleText + " "
	}

	// Calculate left panel width for hover detection
	leftWidth := int(float64(a.width) * 0.30)
	if leftWidth < 20 {
		leftWidth = 20
	}

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
			// Hover: mouse is in left panel and on this item's row
			// Row 0 = title/border, so item i is at row i+1, but we need +1 more for border offset
			hovered := a.mouseX < leftWidth && a.mouseY == i+2
			name := truncateString(wf.Name, width-6)
			content = append(content, a.renderListItem(name, selected, focused, hovered))
		}
	}

	innerWidth := width - 2
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

// buildRunsPanel builds the runs panel for the left sidebar
func (a *App) buildRunsPanel(width, height int) []string {
	if height < 2 {
		return []string{}
	}
	lines := make([]string, height)

	focused := a.focusedPane == RunsPane
	borderColor := UnfocusedColor
	if focused {
		borderColor = FocusedColor
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Title with lazydocker-style inverted colors when focused
	titleText := "Runs"
	var title string
	if focused {
		title = " " + FocusedTitle.Render(" "+titleText+" ") + " "
	} else {
		title = " " + titleText + " "
	}

	// Calculate panel position and left width for hover detection
	leftWidth := int(float64(a.width) * 0.30)
	if leftWidth < 20 {
		leftWidth = 20
	}
	totalHeight := a.height - 1
	if totalHeight < 10 {
		totalHeight = 10
	}
	panelHeight := totalHeight / 3
	panelStartY := panelHeight // Runs panel starts after Workflows panel

	var content []string
	items := a.runs.Items()
	if len(items) == 0 {
		content = append(content, "  Select workflow")
	} else {
		for i, run := range items {
			selected := i == a.runs.SelectedIndex()
			// Hover: mouse is in left panel and on this item's row
			hovered := a.mouseX < leftWidth && a.mouseY == panelStartY+i+2
			icon := StatusIcon(run.Status, run.Conclusion)
			line := icon + " #" + formatRunNumber(run.ID) + " " + run.Branch
			line = truncateString(line, width-6)
			content = append(content, a.renderListItem(line, selected, focused, hovered))
		}
	}

	innerWidth := width - 2
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

// buildJobsPanel builds the jobs panel for the left sidebar
func (a *App) buildJobsPanel(width, height int) []string {
	if height < 2 {
		return []string{}
	}
	lines := make([]string, height)

	focused := a.focusedPane == JobsPane
	borderColor := UnfocusedColor
	if focused {
		borderColor = FocusedColor
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Title with lazydocker-style inverted colors when focused
	titleText := "Jobs"
	var title string
	if focused {
		title = " " + FocusedTitle.Render(" "+titleText+" ") + " "
	} else {
		title = " " + titleText + " "
	}

	// Calculate panel position and left width for hover detection
	leftWidth := int(float64(a.width) * 0.30)
	if leftWidth < 20 {
		leftWidth = 20
	}
	totalHeight := a.height - 1
	if totalHeight < 10 {
		totalHeight = 10
	}
	panelHeight := totalHeight / 3
	panelStartY := 2 * panelHeight // Jobs panel starts after Workflows and Runs panels

	var content []string
	items := a.jobs.Items()
	if len(items) == 0 {
		content = append(content, "  Select a run")
	} else {
		for i, job := range items {
			selected := i == a.jobs.SelectedIndex()
			// Hover: mouse is in left panel and on this item's row
			hovered := a.mouseX < leftWidth && a.mouseY == panelStartY+i+2
			icon := StatusIcon(job.Status, job.Conclusion)
			line := icon + " " + truncateString(job.Name, width-10)
			content = append(content, a.renderListItem(line, selected, focused, hovered))
		}
	}

	innerWidth := width - 2
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

// buildDetailPanel builds the detail view panel (right side) with tabs
func (a *App) buildDetailPanel(width, height int) []string {
	if height < 2 {
		return []string{}
	}
	lines := make([]string, height)

	borderColor := UnfocusedColor
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Tab header
	infoTab := " Info "
	logsTab := " Logs "
	if a.detailTab == InfoTab {
		infoTab = FocusedTitle.Render(" Info ")
	} else {
		logsTab = FocusedTitle.Render(" Logs ")
	}
	tabHeader := " [1]" + infoTab + " [2]" + logsTab + " "

	var content []string

	if a.detailTab == InfoTab {
		content = a.buildInfoContent(width - 4)
	} else {
		content = a.buildLogsContent(width - 4)
	}

	innerWidth := width - 2
	lines[0] = borderStyle.Render("┌") + borderStyle.Render(padCenter(tabHeader, innerWidth, "─")) + borderStyle.Render("┐")
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

// buildLogsContentForDetail builds the content for the Logs tab
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

// padCenter centers a string within the specified width, filling with the given char
func padCenter(s string, width int, fill string) string {
	sLen := lipgloss.Width(s)
	if sLen >= width {
		return s
	}
	leftPad := (width - sLen) / 2
	rightPad := width - sLen - leftPad
	return strings.Repeat(fill, leftPad) + s + strings.Repeat(fill, rightPad)
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
		return LogsLoadedMsg{JobID: job.ID, Logs: logs, Err: err}
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
		maxWidth = 80
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

// Run starts the TUI application
func Run(client github.Client, repo github.Repository) error {
	app := New(
		WithClient(client),
		WithRepository(repo),
	)

	p := tea.NewProgram(app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
