// Package integration provides integration tests for lazyactions TUI application.
package integration

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/lazyactions/app"
	"github.com/nnnkkk7/lazyactions/github"
)

// MockClipboard is a mock clipboard for testing.
type MockClipboard struct {
	Content string
}

// WriteAll implements app.Clipboard.
func (m *MockClipboard) WriteAll(text string) error {
	m.Content = text
	return nil
}

type mockState struct {
	workflows []github.Workflow
	runs      []github.Run
	jobs      []github.Job
	logs      string
	err       error
	rateLimit int
}

func newMockClient(state *mockState) *github.MockClient {
	if state == nil {
		state = &mockState{}
	}

	return &github.MockClient{
		ListWorkflowsFunc: func(ctx context.Context, repo github.Repository) ([]github.Workflow, error) {
			return state.workflows, state.err
		},
		ListRunsFunc: func(ctx context.Context, repo github.Repository, opts *github.ListRunsOpts) ([]github.Run, error) {
			return state.runs, state.err
		},
		ListJobsFunc: func(ctx context.Context, repo github.Repository, runID int64) ([]github.Job, error) {
			return state.jobs, state.err
		},
		GetJobLogsFunc: func(ctx context.Context, repo github.Repository, jobID int64) (string, error) {
			return state.logs, state.err
		},
		CancelRunFunc: func(ctx context.Context, repo github.Repository, runID int64) error {
			return state.err
		},
		RerunWorkflowFunc: func(ctx context.Context, repo github.Repository, runID int64) error {
			return state.err
		},
		RerunFailedJobsFunc: func(ctx context.Context, repo github.Repository, runID int64) error {
			return state.err
		},
		TriggerWorkflowFunc: func(ctx context.Context, repo github.Repository, workflowFile, ref string, inputs map[string]interface{}) error {
			return state.err
		},
		RateLimitRemainingFunc: func() int {
			if state.rateLimit > 0 {
				return state.rateLimit
			}
			return 5000
		},
	}
}

// TestApp wraps App for E2E testing with helper methods.
type TestApp struct {
	*app.App
	t         *testing.T
	mock      *github.MockClient
	mockState *mockState
	clipboard *MockClipboard
}

// TestOption configures a TestApp.
type TestOption func(*TestApp)

// NewTestApp creates a new test application with mock client.
func NewTestApp(t *testing.T, opts ...TestOption) *TestApp {
	t.Helper()

	state := &mockState{rateLimit: 5000}
	mock := newMockClient(state)
	mockClipboard := &MockClipboard{}

	ta := &TestApp{
		App: app.New(
			app.WithClient(mock),
			app.WithRepository(github.Repository{Owner: "test", Name: "repo"}),
			app.WithClipboard(mockClipboard),
		),
		t:         t,
		mock:      mock,
		mockState: state,
		clipboard: mockClipboard,
	}

	for _, opt := range opts {
		opt(ta)
	}

	return ta
}

// WithMockWorkflows sets mock workflows.
func WithMockWorkflows(workflows []github.Workflow) TestOption {
	return func(ta *TestApp) {
		ta.mockState.workflows = workflows
	}
}

// WithMockRuns sets mock runs.
func WithMockRuns(runs []github.Run) TestOption {
	return func(ta *TestApp) {
		ta.mockState.runs = runs
	}
}

// WithMockJobs sets mock jobs.
func WithMockJobs(jobs []github.Job) TestOption {
	return func(ta *TestApp) {
		ta.mockState.jobs = jobs
	}
}

// WithMockLogs sets mock logs.
func WithMockLogs(logs string) TestOption {
	return func(ta *TestApp) {
		ta.mockState.logs = logs
	}
}

// WithMockError sets mock error.
func WithMockError(err error) TestOption {
	return func(ta *TestApp) {
		ta.mockState.err = err
	}
}

// WithMockRateLimit sets mock rate limit.
func WithMockRateLimit(remaining int) TestOption {
	return func(ta *TestApp) {
		ta.mockState.rateLimit = remaining
	}
}

// Mock returns the underlying mock client.
func (ta *TestApp) Mock() *github.MockClient {
	return ta.mock
}

// SendKey simulates a key press and returns the resulting command.
func (ta *TestApp) SendKey(key string) tea.Cmd {
	ta.t.Helper()
	msg := keyMsgFromString(key)
	_, cmd := ta.App.Update(msg)
	return cmd
}

// SendKeyMsg sends a tea.KeyMsg directly.
func (ta *TestApp) SendKeyMsg(msg tea.KeyMsg) tea.Cmd {
	ta.t.Helper()
	_, cmd := ta.App.Update(msg)
	return cmd
}

// SendWindowSize simulates a window resize.
func (ta *TestApp) SendWindowSize(width, height int) {
	ta.t.Helper()
	msg := tea.WindowSizeMsg{Width: width, Height: height}
	ta.App.Update(msg)
}

// ProcessCmd executes a tea.Cmd and returns the resulting message.
func (ta *TestApp) ProcessCmd(cmd tea.Cmd) tea.Msg {
	ta.t.Helper()
	if cmd == nil {
		return nil
	}
	return cmd()
}

// ProcessCmdAndUpdate executes a cmd, sends the result to Update, and returns the next cmd.
func (ta *TestApp) ProcessCmdAndUpdate(cmd tea.Cmd) tea.Cmd {
	ta.t.Helper()
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if msg == nil {
		return nil
	}
	_, nextCmd := ta.App.Update(msg)
	return nextCmd
}

// ProcessCmdChain processes cascading commands up to maxDepth.
func (ta *TestApp) ProcessCmdChain(cmd tea.Cmd, maxDepth int) []tea.Msg {
	ta.t.Helper()
	var msgs []tea.Msg
	for i := 0; i < maxDepth && cmd != nil; i++ {
		msg := cmd()
		if msg == nil {
			break
		}
		msgs = append(msgs, msg)
		_, cmd = ta.App.Update(msg)
	}
	return msgs
}

// SetSize sets the app size for rendering.
func (ta *TestApp) SetSize(width, height int) {
	ta.t.Helper()
	ta.SendWindowSize(width, height)
}

// InitApp initializes the app and processes the init commands.
func (ta *TestApp) InitApp() {
	ta.t.Helper()
	ta.SetSize(120, 40)
	cmd := ta.App.Init()
	ta.ProcessCmdChain(cmd, 10)
}

// ErrTest is a common test error for integration tests.
var ErrTest = &github.AppError{
	Type:    github.ErrTypeNetwork,
	Message: "test error",
}

// keyMsgFromString converts a key string to tea.KeyMsg.
func keyMsgFromString(key string) tea.KeyMsg {
	switch key {
	case "q":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	case "?":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	case "j":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	case "k":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	case "h":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	case "l":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	case "c":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	case "r":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	case "R":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
	case "t":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	case "y":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	case "Y":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}}
	case "n":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	case "L":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}
	case "/":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "ctrl+r":
		return tea.KeyMsg{Type: tea.KeyCtrlR}
	default:
		// Single character
		if len(key) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		return tea.KeyMsg{}
	}
}
