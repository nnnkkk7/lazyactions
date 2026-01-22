// Package app provides the TUI application for lazyactions.
package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

// Layout constants
const (
	// LeftPanelWidthRatio is the percentage of screen width for the left sidebar
	LeftPanelWidthRatio = 0.30
	// LogPaneWidthRatio is the percentage of screen width for the log pane
	LogPaneWidthRatio = 0.50
	// WorkflowsPaneWidthRatio is the percentage of screen width for workflows pane
	WorkflowsPaneWidthRatio = 0.20
	// MinLeftPanelWidth is the minimum width for the left panel
	MinLeftPanelWidth = 20
	// MinTotalHeight is the minimum terminal height
	MinTotalHeight = 10
	// NumLeftPanels is the number of panels in the left sidebar
	NumLeftPanels = 3
	// MinPanelHeight is the minimum height for each panel
	MinPanelHeight = 5
	// FilterInputCharLimit is the maximum characters for the filter input
	FilterInputCharLimit = 50
	// MinLogPaneWidth is the minimum width for log pane
	MinLogPaneWidth = 20
	// MinWorkflowsPaneWidth is the minimum width for workflows pane
	MinWorkflowsPaneWidth = 15
	// MinRunsPaneWidth is the minimum width for runs pane
	MinRunsPaneWidth = 20
	// DefaultLogViewportWidth is the default width for the log viewport
	DefaultLogViewportWidth = 80
	// DefaultLogViewportHeight is the default height for the log viewport
	DefaultLogViewportHeight = 20
	// DefaultWrapWidth is the default width for line wrapping
	DefaultWrapWidth = 80
	// BorderOffset accounts for border lines in panel layout
	BorderOffset = 2
	// BorderWidth is the width taken by left and right borders
	BorderWidth = 2
	// StatusBarHeight is the height of the status bar
	StatusBarHeight = 1
	// ItemPaddingSmall is the padding for truncated item names
	ItemPaddingSmall = 6
	// ItemPaddingMedium is the padding for truncated job names
	ItemPaddingMedium = 10
	// ContentPadding is the padding for content areas
	ContentPadding = 4
	// StatusAreaHeight accounts for status bar and bottom border
	StatusAreaHeight = 2
	// FlashDurationSuccess is the flash message duration for success
	FlashDurationSuccess = 2 * time.Second
	// FlashDurationInfo is the flash message duration for info messages
	FlashDurationInfo = 3 * time.Second
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
	ti.CharLimit = FilterInputCharLimit

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
		logView:         NewLogViewport(DefaultLogViewportWidth, DefaultLogViewportHeight),
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
			if job, ok := a.jobs.Selected(); ok {
				// GitHub API only provides logs for completed jobs
				if job.IsCompleted() && a.parsedLogs == nil {
					cmds = append(cmds, a.fetchLogsCmd(job.ID))
				} else if !job.IsCompleted() {
					a.logView.SetContent(jobStatusMessage(job))
				}
			}
		}

	case LogsLoadedMsg:
		// Only update logs if they are for the currently selected job
		// This prevents stale logs from overwriting newer ones
		job, ok := a.jobs.Selected()
		if !ok || job.ID != msg.JobID {
			break
		}

		if msg.Err != nil {
			a.parsedLogs = nil
			// Don't show error for incomplete jobs - logs aren't available yet
			if job.IsCompleted() {
				a.logView.SetContent("Failed to load logs")
			} else {
				a.logView.SetContent("Waiting for job to complete...")
			}
			// Don't set a.err - avoid showing error in status bar
		} else {
			a.parsedLogs = ParseLogs(msg.Logs)
			a.updateLogViewContent()
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

	// Calculate dimensions using helper
	totalHeight, panelHeight := a.panelLayout()

	// Left sidebar, Right detail
	leftWidth := a.leftPanelWidth()
	rightWidth := a.width - leftWidth

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

// formatRunNumber formats a run ID for display
func formatRunNumber(id int64) string {
	return strconv.FormatInt(id, 10)
}

// Run starts the TUI application
func Run(client github.Client, repo github.Repository) error {
	// Display startup banner
	PrintBanner()

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
