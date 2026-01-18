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
			a.logView.SetContent(msg.Logs)
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

	// Main 3-pane layout
	workflowsPane := a.renderWorkflowsPane()
	runsPane := a.renderRunsPane()
	logsPane := a.renderLogsPane()

	main := lipgloss.JoinHorizontal(lipgloss.Top,
		workflowsPane,
		runsPane,
		logsPane,
	)

	statusBar := a.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, main, statusBar)
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

// Rendering helpers
func (a *App) logPaneWidth() int {
	return int(float64(a.width) * 0.55)
}

func (a *App) logPaneHeight() int {
	return a.height - 2 // account for status bar
}

func (a *App) workflowsPaneWidth() int {
	return int(float64(a.width) * 0.20)
}

func (a *App) runsPaneWidth() int {
	return a.width - a.workflowsPaneWidth() - a.logPaneWidth()
}

func (a *App) paneHeight() int {
	return a.height - 2 // account for status bar
}

func (a *App) renderWorkflowsPane() string {
	style := UnfocusedPane
	titleStyle := UnfocusedTitle
	if a.focusedPane == WorkflowsPane {
		style = FocusedPane
		titleStyle = FocusedTitle
	}

	title := titleStyle.Render("Workflows")
	if a.loading && a.focusedPane == WorkflowsPane {
		title = titleStyle.Render("Workflows " + a.spinner.View())
	}

	var content strings.Builder
	content.WriteString(title + "\n")

	items := a.workflows.Items()
	for i, wf := range items {
		selected := i == a.workflows.SelectedIndex()
		content.WriteString(RenderItem(wf.Name, selected) + "\n")
	}

	scrollPos := ScrollPosition(a.workflows.SelectedIndex(), a.workflows.Len())
	content.WriteString("\n" + scrollPos)

	return style.
		Width(a.workflowsPaneWidth()).
		Height(a.paneHeight()).
		Render(content.String())
}

func (a *App) renderRunsPane() string {
	style := UnfocusedPane
	titleStyle := UnfocusedTitle
	if a.focusedPane == RunsPane {
		style = FocusedPane
		titleStyle = FocusedTitle
	}

	title := titleStyle.Render("Runs")
	if a.loading && a.focusedPane == RunsPane {
		title = titleStyle.Render("Runs " + a.spinner.View())
	}

	var content strings.Builder
	content.WriteString(title + "\n")

	items := a.runs.Items()
	for i, run := range items {
		selected := i == a.runs.SelectedIndex()
		icon := StatusIcon(run.Status, run.Conclusion)
		line := icon + " #" + formatRunNumber(run.ID) + " " + run.Branch
		content.WriteString(RenderItem(line, selected) + "\n")
	}

	scrollPos := ScrollPosition(a.runs.SelectedIndex(), a.runs.Len())
	content.WriteString("\n" + scrollPos)

	return style.
		Width(a.runsPaneWidth()).
		Height(a.paneHeight()).
		Render(content.String())
}

func (a *App) renderLogsPane() string {
	style := UnfocusedPane
	titleStyle := UnfocusedTitle
	if a.focusedPane == LogsPane {
		style = FocusedPane
		titleStyle = FocusedTitle
	}

	title := titleStyle.Render("Logs")

	var content strings.Builder
	content.WriteString(title + "\n")

	// Show job list
	items := a.jobs.Items()
	for i, job := range items {
		selected := i == a.jobs.SelectedIndex()
		icon := StatusIcon(job.Status, job.Conclusion)
		line := icon + " " + job.Name
		content.WriteString(RenderItem(line, selected) + "\n")
	}

	content.WriteString("\n")
	content.WriteString(a.logView.View())

	return style.
		Width(a.logPaneWidth()).
		Height(a.paneHeight()).
		Render(content.String())
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
