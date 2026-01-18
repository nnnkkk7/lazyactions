package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickerTask manages periodic polling tasks.
// It uses a context-based cancellation pattern inspired by lazydocker.
type TickerTask struct {
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration
	taskFn   func(ctx context.Context) tea.Msg
}

// NewTickerTask creates a new TickerTask with the specified interval and task function.
// The task function is called periodically and can return a tea.Msg to stop polling,
// or nil to continue polling.
func NewTickerTask(interval time.Duration, fn func(ctx context.Context) tea.Msg) *TickerTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &TickerTask{
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
		taskFn:   fn,
	}
}

// Start begins the periodic polling and returns a tea.Cmd that runs the polling loop.
// The returned command will:
// - Wait for each tick interval
// - Call the task function with the context
// - If the task function returns non-nil, return that message and stop
// - If the task function returns nil, continue polling
// - If the context is cancelled (via Stop), return nil and exit
func (t *TickerTask) Start() tea.Cmd {
	return func() tea.Msg {
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		for {
			select {
			case <-t.ctx.Done():
				return nil
			case <-ticker.C:
				if msg := t.taskFn(t.ctx); msg != nil {
					return msg
				}
			}
		}
	}
}

// Stop cancels the polling loop.
// It is safe to call Stop multiple times.
// After Stop is called, the Start command will exit and return nil.
func (t *TickerTask) Stop() {
	t.cancel()
}
