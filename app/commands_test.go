package app

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/lazyactions/github"
)

func TestFetchWorkflows(t *testing.T) {
	t.Run("returns workflows on success", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			workflows: []github.Workflow{
				{ID: 1, Name: "CI", Path: ".github/workflows/ci.yml", State: "active"},
				{ID: 2, Name: "Deploy", Path: ".github/workflows/deploy.yml", State: "active"},
			},
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}

		cmd := fetchWorkflows(mock, repo)
		msg := cmd()

		result, ok := msg.(WorkflowsLoadedMsg)
		if !ok {
			t.Fatalf("expected WorkflowsLoadedMsg, got %T", msg)
		}
		if len(result.Workflows) != 2 {
			t.Errorf("expected 2 workflows, got %d", len(result.Workflows))
		}
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
		if len(mock.ListWorkflowsCalls()) != 1 {
			t.Errorf("expected 1 call to ListWorkflows, got %d", len(mock.ListWorkflowsCalls()))
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			err: errors.New("API error"),
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}

		cmd := fetchWorkflows(mock, repo)
		msg := cmd()

		result, ok := msg.(WorkflowsLoadedMsg)
		if !ok {
			t.Fatalf("expected WorkflowsLoadedMsg, got %T", msg)
		}
		if result.Err == nil {
			t.Error("expected error, got nil")
		}
		if result.Err.Error() != "API error" {
			t.Errorf("expected 'API error', got %q", result.Err.Error())
		}
	})
}

func TestFetchRuns(t *testing.T) {
	t.Run("returns runs on success", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			runs: []github.Run{
				{ID: 100, Name: "CI", Status: "completed", Conclusion: "success"},
				{ID: 101, Name: "CI", Status: "in_progress", Conclusion: ""},
			},
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}
		workflowID := int64(1)

		cmd := fetchRuns(mock, repo, workflowID)
		msg := cmd()

		result, ok := msg.(RunsLoadedMsg)
		if !ok {
			t.Fatalf("expected RunsLoadedMsg, got %T", msg)
		}
		if len(result.Runs) != 2 {
			t.Errorf("expected 2 runs, got %d", len(result.Runs))
		}
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
		if len(mock.ListRunsCalls()) != 1 {
			t.Errorf("expected 1 call to ListRuns, got %d", len(mock.ListRunsCalls()))
		}

		// Verify the workflow ID was passed in opts
		calls := mock.ListRunsCalls()
		if len(calls) == 0 {
			t.Fatal("expected calls to be recorded")
		}
		opts := calls[0].Opts
		if opts.WorkflowID != workflowID {
			t.Errorf("expected workflow ID %d, got %d", workflowID, opts.WorkflowID)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			err: errors.New("network error"),
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}

		cmd := fetchRuns(mock, repo, 1)
		msg := cmd()

		result, ok := msg.(RunsLoadedMsg)
		if !ok {
			t.Fatalf("expected RunsLoadedMsg, got %T", msg)
		}
		if result.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestFetchJobs(t *testing.T) {
	t.Run("returns jobs on success", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			jobs: []github.Job{
				{ID: 200, Name: "build", Status: "completed", Conclusion: "success"},
				{ID: 201, Name: "test", Status: "completed", Conclusion: "failure"},
			},
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}
		runID := int64(100)

		cmd := fetchJobs(mock, repo, runID)
		msg := cmd()

		result, ok := msg.(JobsLoadedMsg)
		if !ok {
			t.Fatalf("expected JobsLoadedMsg, got %T", msg)
		}
		if len(result.Jobs) != 2 {
			t.Errorf("expected 2 jobs, got %d", len(result.Jobs))
		}
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
		if len(mock.ListJobsCalls()) != 1 {
			t.Errorf("expected 1 call to ListJobs, got %d", len(mock.ListJobsCalls()))
		}

		// Verify the run ID was passed
		calls := mock.ListJobsCalls()
		if len(calls) == 0 {
			t.Fatal("expected calls to be recorded")
		}
		passedRunID := calls[0].RunID
		if passedRunID != runID {
			t.Errorf("expected run ID %d, got %d", runID, passedRunID)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			err: errors.New("job fetch failed"),
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}

		cmd := fetchJobs(mock, repo, 100)
		msg := cmd()

		result, ok := msg.(JobsLoadedMsg)
		if !ok {
			t.Fatalf("expected JobsLoadedMsg, got %T", msg)
		}
		if result.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestFetchLogs(t *testing.T) {
	t.Run("returns logs on success", func(t *testing.T) {
		expectedLogs := "2024-01-01T00:00:00Z Running tests...\n2024-01-01T00:00:01Z Tests passed!"
		mock := newMockClient(&mockClientState{
			logs: expectedLogs,
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}
		jobID := int64(200)

		cmd := fetchLogs(mock, repo, jobID)
		msg := cmd()

		result, ok := msg.(LogsLoadedMsg)
		if !ok {
			t.Fatalf("expected LogsLoadedMsg, got %T", msg)
		}
		if result.Logs != expectedLogs {
			t.Errorf("expected logs %q, got %q", expectedLogs, result.Logs)
		}
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
		if len(mock.GetJobLogsCalls()) != 1 {
			t.Errorf("expected 1 call to GetJobLogs, got %d", len(mock.GetJobLogsCalls()))
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			err: errors.New("logs not available"),
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}

		cmd := fetchLogs(mock, repo, 200)
		msg := cmd()

		result, ok := msg.(LogsLoadedMsg)
		if !ok {
			t.Fatalf("expected LogsLoadedMsg, got %T", msg)
		}
		if result.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestCancelRun(t *testing.T) {
	t.Run("cancels run successfully", func(t *testing.T) {
		mock := newMockClient(nil)
		repo := github.Repository{Owner: "owner", Name: "repo"}
		runID := int64(12345)

		cmd := cancelRun(mock, repo, runID)
		msg := cmd()

		result, ok := msg.(RunCancelledMsg)
		if !ok {
			t.Fatalf("expected RunCancelledMsg, got %T", msg)
		}
		if result.RunID != runID {
			t.Errorf("expected run ID %d, got %d", runID, result.RunID)
		}
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
		if len(mock.CancelRunCalls()) != 1 {
			t.Errorf("expected 1 call to CancelRun, got %d", len(mock.CancelRunCalls()))
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			err: errors.New("cannot cancel completed run"),
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}
		runID := int64(12345)

		cmd := cancelRun(mock, repo, runID)
		msg := cmd()

		result, ok := msg.(RunCancelledMsg)
		if !ok {
			t.Fatalf("expected RunCancelledMsg, got %T", msg)
		}
		if result.RunID != runID {
			t.Errorf("expected run ID %d, got %d", runID, result.RunID)
		}
		if result.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRerunWorkflow(t *testing.T) {
	t.Run("reruns workflow successfully", func(t *testing.T) {
		mock := newMockClient(nil)
		repo := github.Repository{Owner: "owner", Name: "repo"}
		runID := int64(67890)

		cmd := rerunWorkflow(mock, repo, runID)
		msg := cmd()

		result, ok := msg.(RunRerunMsg)
		if !ok {
			t.Fatalf("expected RunRerunMsg, got %T", msg)
		}
		if result.RunID != runID {
			t.Errorf("expected run ID %d, got %d", runID, result.RunID)
		}
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
		if len(mock.RerunWorkflowCalls()) != 1 {
			t.Errorf("expected 1 call to RerunWorkflow, got %d", len(mock.RerunWorkflowCalls()))
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			err: errors.New("rerun failed"),
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}
		runID := int64(67890)

		cmd := rerunWorkflow(mock, repo, runID)
		msg := cmd()

		result, ok := msg.(RunRerunMsg)
		if !ok {
			t.Fatalf("expected RunRerunMsg, got %T", msg)
		}
		if result.RunID != runID {
			t.Errorf("expected run ID %d, got %d", runID, result.RunID)
		}
		if result.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRerunFailedJobs(t *testing.T) {
	t.Run("reruns failed jobs successfully", func(t *testing.T) {
		mock := newMockClient(nil)
		repo := github.Repository{Owner: "owner", Name: "repo"}
		runID := int64(11111)

		cmd := rerunFailedJobs(mock, repo, runID)
		msg := cmd()

		result, ok := msg.(RerunFailedJobsMsg)
		if !ok {
			t.Fatalf("expected RerunFailedJobsMsg, got %T", msg)
		}
		if result.RunID != runID {
			t.Errorf("expected run ID %d, got %d", runID, result.RunID)
		}
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
		if len(mock.RerunFailedJobsCalls()) != 1 {
			t.Errorf("expected 1 call to RerunFailedJobs, got %d", len(mock.RerunFailedJobsCalls()))
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			err: errors.New("no failed jobs to rerun"),
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}
		runID := int64(11111)

		cmd := rerunFailedJobs(mock, repo, runID)
		msg := cmd()

		result, ok := msg.(RerunFailedJobsMsg)
		if !ok {
			t.Fatalf("expected RerunFailedJobsMsg, got %T", msg)
		}
		if result.RunID != runID {
			t.Errorf("expected run ID %d, got %d", runID, result.RunID)
		}
		if result.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestTriggerWorkflow(t *testing.T) {
	t.Run("triggers workflow successfully", func(t *testing.T) {
		mock := newMockClient(nil)
		repo := github.Repository{Owner: "owner", Name: "repo"}
		workflowFile := "deploy.yml"
		ref := "main"
		inputs := map[string]string{"environment": "production"}

		cmd := triggerWorkflow(mock, repo, workflowFile, ref, inputs)
		msg := cmd()

		result, ok := msg.(WorkflowTriggeredMsg)
		if !ok {
			t.Fatalf("expected WorkflowTriggeredMsg, got %T", msg)
		}
		if result.Workflow != workflowFile {
			t.Errorf("expected workflow %q, got %q", workflowFile, result.Workflow)
		}
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
		if len(mock.TriggerWorkflowCalls()) != 1 {
			t.Errorf("expected 1 call to TriggerWorkflow, got %d", len(mock.TriggerWorkflowCalls()))
		}

		// Verify the inputs were passed correctly
		calls := mock.TriggerWorkflowCalls()
		if len(calls) == 0 {
			t.Fatal("expected calls to be recorded")
		}
		passedInputs := calls[0].Inputs
		if passedInputs["environment"] != "production" {
			t.Errorf("expected input 'environment' to be 'production', got %v", passedInputs["environment"])
		}
	})

	t.Run("triggers workflow with nil inputs", func(t *testing.T) {
		mock := newMockClient(nil)
		repo := github.Repository{Owner: "owner", Name: "repo"}

		cmd := triggerWorkflow(mock, repo, "ci.yml", "main", nil)
		msg := cmd()

		result, ok := msg.(WorkflowTriggeredMsg)
		if !ok {
			t.Fatalf("expected WorkflowTriggeredMsg, got %T", msg)
		}
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := newMockClient(&mockClientState{
			err: errors.New("workflow not found"),
		})
		repo := github.Repository{Owner: "owner", Name: "repo"}

		cmd := triggerWorkflow(mock, repo, "missing.yml", "main", nil)
		msg := cmd()

		result, ok := msg.(WorkflowTriggeredMsg)
		if !ok {
			t.Fatalf("expected WorkflowTriggeredMsg, got %T", msg)
		}
		if result.Workflow != "missing.yml" {
			t.Errorf("expected workflow 'missing.yml', got %q", result.Workflow)
		}
		if result.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestFlashMessage(t *testing.T) {
	t.Run("returns batch command", func(t *testing.T) {
		duration := 3 * time.Second
		cmd := flashMessage("Operation successful!", duration)

		// The command should not be nil
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}

		// Execute the batch command - it returns a batch message
		// We can't easily test the internal behavior of tea.Batch,
		// but we can verify it's a valid command
		msg := cmd()
		if msg == nil {
			t.Error("expected non-nil message from batch command")
		}
	})
}

func TestTick(t *testing.T) {
	t.Run("creates tick command", func(t *testing.T) {
		interval := 10 * time.Second
		cmd := tick(interval)

		// The command should not be nil
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}

		// Note: We can't easily test the timing behavior without
		// actually waiting, but we've verified the command is created
	})
}

// TestCommandsAreFunctions verifies that all commands return tea.Cmd types
func TestCommandsAreFunctions(t *testing.T) {
	mock := newMockClient(nil)
	repo := github.Repository{Owner: "owner", Name: "repo"}

	// All these should compile and return tea.Cmd
	var cmd tea.Cmd

	cmd = fetchWorkflows(mock, repo)
	if cmd == nil {
		t.Error("fetchWorkflows returned nil")
	}

	cmd = fetchRuns(mock, repo, 1)
	if cmd == nil {
		t.Error("fetchRuns returned nil")
	}

	cmd = fetchJobs(mock, repo, 1)
	if cmd == nil {
		t.Error("fetchJobs returned nil")
	}

	cmd = fetchLogs(mock, repo, 1)
	if cmd == nil {
		t.Error("fetchLogs returned nil")
	}

	cmd = cancelRun(mock, repo, 1)
	if cmd == nil {
		t.Error("cancelRun returned nil")
	}

	cmd = rerunWorkflow(mock, repo, 1)
	if cmd == nil {
		t.Error("rerunWorkflow returned nil")
	}

	cmd = rerunFailedJobs(mock, repo, 1)
	if cmd == nil {
		t.Error("rerunFailedJobs returned nil")
	}

	cmd = triggerWorkflow(mock, repo, "ci.yml", "main", nil)
	if cmd == nil {
		t.Error("triggerWorkflow returned nil")
	}

	cmd = flashMessage("test", time.Second)
	if cmd == nil {
		t.Error("flashMessage returned nil")
	}

	cmd = tick(time.Second)
	if cmd == nil {
		t.Error("tick returned nil")
	}
}

// TestConcurrentCommandExecution verifies commands can be safely executed concurrently
func TestConcurrentCommandExecution(t *testing.T) {
	mock := newMockClient(&mockClientState{
		workflows: []github.Workflow{{ID: 1, Name: "CI"}},
		runs:      []github.Run{{ID: 100, Name: "CI"}},
		jobs:      []github.Job{{ID: 200, Name: "build"}},
		logs:      "test logs",
	})
	repo := github.Repository{Owner: "owner", Name: "repo"}

	// Create multiple commands
	cmds := []tea.Cmd{
		fetchWorkflows(mock, repo),
		fetchRuns(mock, repo, 1),
		fetchJobs(mock, repo, 100),
		fetchLogs(mock, repo, 200),
	}

	// Execute them concurrently
	done := make(chan tea.Msg, len(cmds))
	for _, cmd := range cmds {
		go func(c tea.Cmd) {
			done <- c()
		}(cmd)
	}

	// Collect all results
	results := make([]tea.Msg, 0, len(cmds))
	for i := 0; i < len(cmds); i++ {
		results = append(results, <-done)
	}

	if len(results) != len(cmds) {
		t.Errorf("expected %d results, got %d", len(cmds), len(results))
	}

	// Verify all expected message types are present
	hasWorkflows := false
	hasRuns := false
	hasJobs := false
	hasLogs := false

	for _, result := range results {
		switch result.(type) {
		case WorkflowsLoadedMsg:
			hasWorkflows = true
		case RunsLoadedMsg:
			hasRuns = true
		case JobsLoadedMsg:
			hasJobs = true
		case LogsLoadedMsg:
			hasLogs = true
		}
	}

	if !hasWorkflows {
		t.Error("missing WorkflowsLoadedMsg")
	}
	if !hasRuns {
		t.Error("missing RunsLoadedMsg")
	}
	if !hasJobs {
		t.Error("missing JobsLoadedMsg")
	}
	if !hasLogs {
		t.Error("missing LogsLoadedMsg")
	}
}

// TestValueCapture verifies that commands capture values correctly
func TestValueCapture(t *testing.T) {
	mock := newMockClient(nil)
	repo := github.Repository{Owner: "owner", Name: "repo"}

	// Create commands with specific values
	runID := int64(12345)
	cmd := cancelRun(mock, repo, runID)

	// Modify the original variable
	runID = 99999

	// Execute the command - it should still use the captured value
	msg := cmd()
	result, ok := msg.(RunCancelledMsg)
	if !ok {
		t.Fatalf("expected RunCancelledMsg, got %T", msg)
	}

	// The captured run ID should be 12345, not 99999
	if result.RunID != 12345 {
		t.Errorf("expected run ID 12345, got %d", result.RunID)
	}
}
