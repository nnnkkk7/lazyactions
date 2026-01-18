package github

import (
	"context"
	"sync"
	"time"
)

// MockClient is a test mock with call tracking capabilities.
type MockClient struct {
	// Return values
	Workflows []Workflow
	Runs      []Run
	Jobs      []Job
	Logs      string
	Err       error
	RateLimit int

	// Call tracking
	calls []MethodCall
	mu    sync.Mutex
}

// MethodCall records a method invocation.
type MethodCall struct {
	Method string
	Args   []interface{}
	Time   time.Time
}

// RecordCall records a method call with its arguments.
func (m *MockClient) RecordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, MethodCall{
		Method: method,
		Args:   args,
		Time:   time.Now(),
	})
}

// Calls returns a copy of all recorded calls.
func (m *MockClient) Calls() []MethodCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MethodCall{}, m.calls...)
}

// CallCount returns the number of times a method was called.
func (m *MockClient) CallCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, c := range m.calls {
		if c.Method == method {
			count++
		}
	}
	return count
}

// Reset clears the call history.
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
}

// ListWorkflows implements Client.ListWorkflows.
func (m *MockClient) ListWorkflows(ctx context.Context, repo Repository) ([]Workflow, error) {
	m.RecordCall("ListWorkflows", repo)
	return m.Workflows, m.Err
}

// ListRuns implements Client.ListRuns.
func (m *MockClient) ListRuns(ctx context.Context, repo Repository, opts *ListRunsOpts) ([]Run, error) {
	m.RecordCall("ListRuns", repo, opts)
	return m.Runs, m.Err
}

// ListJobs implements Client.ListJobs.
func (m *MockClient) ListJobs(ctx context.Context, repo Repository, runID int64) ([]Job, error) {
	m.RecordCall("ListJobs", repo, runID)
	return m.Jobs, m.Err
}

// GetJobLogs implements Client.GetJobLogs.
func (m *MockClient) GetJobLogs(ctx context.Context, repo Repository, jobID int64) (string, error) {
	m.RecordCall("GetJobLogs", repo, jobID)
	return m.Logs, m.Err
}

// CancelRun implements Client.CancelRun.
func (m *MockClient) CancelRun(ctx context.Context, repo Repository, runID int64) error {
	m.RecordCall("CancelRun", repo, runID)
	return m.Err
}

// RerunWorkflow implements Client.RerunWorkflow.
func (m *MockClient) RerunWorkflow(ctx context.Context, repo Repository, runID int64) error {
	m.RecordCall("RerunWorkflow", repo, runID)
	return m.Err
}

// RerunFailedJobs implements Client.RerunFailedJobs.
func (m *MockClient) RerunFailedJobs(ctx context.Context, repo Repository, runID int64) error {
	m.RecordCall("RerunFailedJobs", repo, runID)
	return m.Err
}

// TriggerWorkflow implements Client.TriggerWorkflow.
func (m *MockClient) TriggerWorkflow(ctx context.Context, repo Repository, workflowFile, ref string, inputs map[string]interface{}) error {
	m.RecordCall("TriggerWorkflow", repo, workflowFile, ref, inputs)
	return m.Err
}

// RateLimitRemaining implements Client.RateLimitRemaining.
func (m *MockClient) RateLimitRemaining() int {
	m.RecordCall("RateLimitRemaining")
	if m.RateLimit > 0 {
		return m.RateLimit
	}
	return 5000 // Default rate limit
}
