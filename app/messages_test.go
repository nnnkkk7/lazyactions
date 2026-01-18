package app

import (
	"errors"
	"testing"
	"time"

	"github.com/nnnkkk7/lazyactions/github"
)

func TestWorkflowsLoadedMsg(t *testing.T) {
	t.Run("with workflows", func(t *testing.T) {
		workflows := []github.Workflow{
			{ID: 1, Name: "CI", Path: ".github/workflows/ci.yml", State: "active"},
			{ID: 2, Name: "Deploy", Path: ".github/workflows/deploy.yml", State: "active"},
		}
		msg := WorkflowsLoadedMsg{Workflows: workflows, Err: nil}

		if len(msg.Workflows) != 2 {
			t.Errorf("expected 2 workflows, got %d", len(msg.Workflows))
		}
		if msg.Err != nil {
			t.Errorf("expected no error, got %v", msg.Err)
		}
		if msg.Workflows[0].Name != "CI" {
			t.Errorf("expected workflow name 'CI', got %q", msg.Workflows[0].Name)
		}
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("API error")
		msg := WorkflowsLoadedMsg{Workflows: nil, Err: err}

		if msg.Workflows != nil {
			t.Errorf("expected nil workflows, got %v", msg.Workflows)
		}
		if msg.Err == nil {
			t.Error("expected error, got nil")
		}
		if msg.Err.Error() != "API error" {
			t.Errorf("expected 'API error', got %q", msg.Err.Error())
		}
	})
}

func TestRunsLoadedMsg(t *testing.T) {
	t.Run("with runs", func(t *testing.T) {
		runs := []github.Run{
			{ID: 100, Name: "CI", Status: "completed", Conclusion: "success", Branch: "main"},
			{ID: 101, Name: "CI", Status: "in_progress", Conclusion: "", Branch: "feature"},
		}
		msg := RunsLoadedMsg{Runs: runs, Err: nil}

		if len(msg.Runs) != 2 {
			t.Errorf("expected 2 runs, got %d", len(msg.Runs))
		}
		if msg.Err != nil {
			t.Errorf("expected no error, got %v", msg.Err)
		}
		if msg.Runs[0].Status != "completed" {
			t.Errorf("expected status 'completed', got %q", msg.Runs[0].Status)
		}
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("network error")
		msg := RunsLoadedMsg{Runs: nil, Err: err}

		if msg.Runs != nil {
			t.Errorf("expected nil runs, got %v", msg.Runs)
		}
		if msg.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestJobsLoadedMsg(t *testing.T) {
	t.Run("with jobs", func(t *testing.T) {
		jobs := []github.Job{
			{ID: 200, Name: "build", Status: "completed", Conclusion: "success"},
			{ID: 201, Name: "test", Status: "completed", Conclusion: "failure"},
		}
		msg := JobsLoadedMsg{Jobs: jobs, Err: nil}

		if len(msg.Jobs) != 2 {
			t.Errorf("expected 2 jobs, got %d", len(msg.Jobs))
		}
		if msg.Err != nil {
			t.Errorf("expected no error, got %v", msg.Err)
		}
		if msg.Jobs[1].Conclusion != "failure" {
			t.Errorf("expected conclusion 'failure', got %q", msg.Jobs[1].Conclusion)
		}
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("job fetch failed")
		msg := JobsLoadedMsg{Jobs: nil, Err: err}

		if msg.Jobs != nil {
			t.Errorf("expected nil jobs, got %v", msg.Jobs)
		}
		if msg.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestLogsLoadedMsg(t *testing.T) {
	t.Run("with logs", func(t *testing.T) {
		logs := "2024-01-01T00:00:00Z Running tests...\n2024-01-01T00:00:01Z Tests passed!"
		msg := LogsLoadedMsg{Logs: logs, Err: nil}

		if msg.Logs == "" {
			t.Error("expected logs, got empty string")
		}
		if msg.Err != nil {
			t.Errorf("expected no error, got %v", msg.Err)
		}
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("logs not available")
		msg := LogsLoadedMsg{Logs: "", Err: err}

		if msg.Logs != "" {
			t.Errorf("expected empty logs, got %q", msg.Logs)
		}
		if msg.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRunCancelledMsg(t *testing.T) {
	t.Run("successful cancellation", func(t *testing.T) {
		msg := RunCancelledMsg{RunID: 12345, Err: nil}

		if msg.RunID != 12345 {
			t.Errorf("expected run ID 12345, got %d", msg.RunID)
		}
		if msg.Err != nil {
			t.Errorf("expected no error, got %v", msg.Err)
		}
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("cannot cancel completed run")
		msg := RunCancelledMsg{RunID: 12345, Err: err}

		if msg.RunID != 12345 {
			t.Errorf("expected run ID 12345, got %d", msg.RunID)
		}
		if msg.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRunRerunMsg(t *testing.T) {
	t.Run("successful rerun", func(t *testing.T) {
		msg := RunRerunMsg{RunID: 67890, Err: nil}

		if msg.RunID != 67890 {
			t.Errorf("expected run ID 67890, got %d", msg.RunID)
		}
		if msg.Err != nil {
			t.Errorf("expected no error, got %v", msg.Err)
		}
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("rerun failed")
		msg := RunRerunMsg{RunID: 67890, Err: err}

		if msg.RunID != 67890 {
			t.Errorf("expected run ID 67890, got %d", msg.RunID)
		}
		if msg.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRerunFailedJobsMsg(t *testing.T) {
	t.Run("successful rerun", func(t *testing.T) {
		msg := RerunFailedJobsMsg{RunID: 11111, Err: nil}

		if msg.RunID != 11111 {
			t.Errorf("expected run ID 11111, got %d", msg.RunID)
		}
		if msg.Err != nil {
			t.Errorf("expected no error, got %v", msg.Err)
		}
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("no failed jobs to rerun")
		msg := RerunFailedJobsMsg{RunID: 11111, Err: err}

		if msg.RunID != 11111 {
			t.Errorf("expected run ID 11111, got %d", msg.RunID)
		}
		if msg.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestWorkflowTriggeredMsg(t *testing.T) {
	t.Run("successful trigger", func(t *testing.T) {
		msg := WorkflowTriggeredMsg{Workflow: "deploy.yml", Err: nil}

		if msg.Workflow != "deploy.yml" {
			t.Errorf("expected workflow 'deploy.yml', got %q", msg.Workflow)
		}
		if msg.Err != nil {
			t.Errorf("expected no error, got %v", msg.Err)
		}
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("workflow not found")
		msg := WorkflowTriggeredMsg{Workflow: "missing.yml", Err: err}

		if msg.Workflow != "missing.yml" {
			t.Errorf("expected workflow 'missing.yml', got %q", msg.Workflow)
		}
		if msg.Err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestFlashMsg(t *testing.T) {
	t.Run("with message and duration", func(t *testing.T) {
		duration := 3 * time.Second
		msg := FlashMsg{Message: "Operation successful!", Duration: duration}

		if msg.Message != "Operation successful!" {
			t.Errorf("expected message 'Operation successful!', got %q", msg.Message)
		}
		if msg.Duration != duration {
			t.Errorf("expected duration %v, got %v", duration, msg.Duration)
		}
	})

	t.Run("with empty message", func(t *testing.T) {
		msg := FlashMsg{Message: "", Duration: time.Second}

		if msg.Message != "" {
			t.Errorf("expected empty message, got %q", msg.Message)
		}
	})
}

func TestFlashClearMsg(t *testing.T) {
	// FlashClearMsg is a simple struct with no fields
	msg := FlashClearMsg{}
	_ = msg // Verify it can be instantiated
}

func TestTickMsg(t *testing.T) {
	now := time.Now()
	msg := TickMsg{Time: now}

	if msg.Time != now {
		t.Errorf("expected time %v, got %v", now, msg.Time)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	t.Run("with dimensions", func(t *testing.T) {
		msg := WindowSizeMsg{Width: 120, Height: 40}

		if msg.Width != 120 {
			t.Errorf("expected width 120, got %d", msg.Width)
		}
		if msg.Height != 40 {
			t.Errorf("expected height 40, got %d", msg.Height)
		}
	})

	t.Run("with zero dimensions", func(t *testing.T) {
		msg := WindowSizeMsg{Width: 0, Height: 0}

		if msg.Width != 0 {
			t.Errorf("expected width 0, got %d", msg.Width)
		}
		if msg.Height != 0 {
			t.Errorf("expected height 0, got %d", msg.Height)
		}
	})
}
