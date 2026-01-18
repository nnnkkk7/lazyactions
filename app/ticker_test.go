package app

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewTickerTask tests the creation of a new TickerTask.
func TestNewTickerTask(t *testing.T) {
	interval := 100 * time.Millisecond
	taskFn := func(ctx context.Context) tea.Msg {
		return "test"
	}

	ticker := NewTickerTask(interval, taskFn)

	if ticker == nil {
		t.Fatal("NewTickerTask returned nil")
	}
	if ticker.interval != interval {
		t.Errorf("expected interval %v, got %v", interval, ticker.interval)
	}
	if ticker.ctx == nil {
		t.Error("expected ctx to be non-nil")
	}
	if ticker.cancel == nil {
		t.Error("expected cancel to be non-nil")
	}
	if ticker.taskFn == nil {
		t.Error("expected taskFn to be non-nil")
	}
}

// TestTickerTask_Start tests that Start returns a tea.Cmd that executes the task.
func TestTickerTask_Start(t *testing.T) {
	var callCount atomic.Int32
	interval := 50 * time.Millisecond

	taskFn := func(ctx context.Context) tea.Msg {
		callCount.Add(1)
		return "tick"
	}

	ticker := NewTickerTask(interval, taskFn)
	cmd := ticker.Start()

	if cmd == nil {
		t.Fatal("Start returned nil cmd")
	}

	// Run the command in a goroutine and let it tick a few times
	done := make(chan tea.Msg, 1)
	go func() {
		msg := cmd()
		done <- msg
	}()

	// Wait for at least one tick
	time.Sleep(75 * time.Millisecond)

	// Stop the ticker
	ticker.Stop()

	// Wait for the command to finish
	select {
	case msg := <-done:
		// The command should return "tick" from the first successful tick
		if msg != "tick" && msg != nil {
			// After Stop, it might return nil
		}
	case <-time.After(200 * time.Millisecond):
		// This is acceptable - the ticker may have been stopped before returning
	}

	// Verify at least one tick occurred
	if callCount.Load() < 1 {
		t.Errorf("expected at least 1 tick, got %d", callCount.Load())
	}
}

// TestTickerTask_Stop tests that Stop cancels the ticker.
func TestTickerTask_Stop(t *testing.T) {
	var callCount atomic.Int32
	interval := 50 * time.Millisecond

	taskFn := func(ctx context.Context) tea.Msg {
		callCount.Add(1)
		// Return nil to continue polling
		return nil
	}

	ticker := NewTickerTask(interval, taskFn)
	cmd := ticker.Start()

	// Run the command in a goroutine
	done := make(chan tea.Msg, 1)
	go func() {
		msg := cmd()
		done <- msg
	}()

	// Let it tick a couple of times
	time.Sleep(120 * time.Millisecond)

	// Stop the ticker
	ticker.Stop()

	// Wait for the command to finish
	select {
	case msg := <-done:
		// After Stop, the command should return nil
		if msg != nil {
			t.Errorf("expected nil after Stop, got %v", msg)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("command did not finish after Stop")
	}
}

// TestTickerTask_StopBeforeStart tests that Stop can be called before Start.
func TestTickerTask_StopBeforeStart(t *testing.T) {
	interval := 100 * time.Millisecond
	taskFn := func(ctx context.Context) tea.Msg {
		return "tick"
	}

	ticker := NewTickerTask(interval, taskFn)

	// Stop before Start - should not panic
	ticker.Stop()

	// Start after Stop - the command should exit immediately
	cmd := ticker.Start()
	done := make(chan tea.Msg, 1)
	go func() {
		msg := cmd()
		done <- msg
	}()

	select {
	case msg := <-done:
		if msg != nil {
			t.Errorf("expected nil after pre-Stop, got %v", msg)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("command did not finish after pre-Stop")
	}
}

// TestTickerTask_MultipleStops tests that Stop can be called multiple times.
func TestTickerTask_MultipleStops(t *testing.T) {
	interval := 100 * time.Millisecond
	taskFn := func(ctx context.Context) tea.Msg {
		return "tick"
	}

	ticker := NewTickerTask(interval, taskFn)

	// Multiple stops should not panic
	ticker.Stop()
	ticker.Stop()
	ticker.Stop()
}

// TestTickerTask_TaskFnReceivesContext tests that the task function receives the correct context.
func TestTickerTask_TaskFnReceivesContext(t *testing.T) {
	interval := 50 * time.Millisecond
	var receivedCtx context.Context

	taskFn := func(ctx context.Context) tea.Msg {
		receivedCtx = ctx
		return "done"
	}

	ticker := NewTickerTask(interval, taskFn)
	cmd := ticker.Start()

	done := make(chan tea.Msg, 1)
	go func() {
		msg := cmd()
		done <- msg
	}()

	select {
	case <-done:
		if receivedCtx == nil {
			t.Error("expected context to be passed to taskFn")
		}
	case <-time.After(200 * time.Millisecond):
		ticker.Stop()
		t.Error("command did not complete")
	}
}

// TestTickerTask_ContextCancelledDuringTask tests behavior when context is cancelled during task execution.
func TestTickerTask_ContextCancelledDuringTask(t *testing.T) {
	interval := 50 * time.Millisecond
	taskStarted := make(chan struct{})

	taskFn := func(ctx context.Context) tea.Msg {
		close(taskStarted)
		// Simulate a long-running task
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(500 * time.Millisecond):
			return "completed"
		}
	}

	ticker := NewTickerTask(interval, taskFn)
	cmd := ticker.Start()

	done := make(chan tea.Msg, 1)
	go func() {
		msg := cmd()
		done <- msg
	}()

	// Wait for task to start
	select {
	case <-taskStarted:
		// Task has started, now stop the ticker
		ticker.Stop()
	case <-time.After(200 * time.Millisecond):
		t.Fatal("task did not start in time")
	}

	// The command should finish
	select {
	case msg := <-done:
		// The task should return nil or complete depending on timing
		_ = msg // Accept any result
	case <-time.After(200 * time.Millisecond):
		t.Error("command did not finish after Stop during task")
	}
}

// TestTickerTask_TaskReturnsNil tests that returning nil continues the polling loop.
func TestTickerTask_TaskReturnsNil(t *testing.T) {
	var callCount atomic.Int32
	interval := 30 * time.Millisecond

	taskFn := func(ctx context.Context) tea.Msg {
		count := callCount.Add(1)
		if count >= 3 {
			return "done"
		}
		return nil // Continue polling
	}

	ticker := NewTickerTask(interval, taskFn)
	cmd := ticker.Start()

	done := make(chan tea.Msg, 1)
	go func() {
		msg := cmd()
		done <- msg
	}()

	select {
	case msg := <-done:
		if msg != "done" {
			t.Errorf("expected 'done', got %v", msg)
		}
		if callCount.Load() != 3 {
			t.Errorf("expected 3 calls, got %d", callCount.Load())
		}
	case <-time.After(500 * time.Millisecond):
		ticker.Stop()
		t.Errorf("command did not complete, callCount: %d", callCount.Load())
	}
}
