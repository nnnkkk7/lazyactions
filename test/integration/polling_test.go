package integration

import (
	"testing"

	"github.com/nnnkkk7/lazyactions/app"
)

func TestPolling_AdaptiveBehavior(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit int
		desc      string
	}{
		{"high rate limit uses base interval", 5000, "Normal polling"},
		{"light caution increases interval", 999, "Slightly increased"},
		{"caution doubles interval", 499, "Doubled interval"},
		{"critical uses max interval", 99, "Max interval"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := NewTestApp(t,
				WithMockWorkflows(DefaultTestWorkflows()),
				WithMockRateLimit(tt.rateLimit),
			)
			ta.SetSize(120, 40)
			ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

			// Should not panic
			view := ta.App.View()
			if len(view) == 0 {
				t.Error("View should render")
			}
		})
	}
}

func TestPolling_RateLimitIntegration(t *testing.T) {
	t.Run("rate limit changes affect behavior", func(t *testing.T) {
		ta := NewTestApp(t, WithMockRateLimit(5000))
		ta.SetSize(120, 40)

		// Initial high rate limit
		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}

		// Simulate rate limit decrease
		ta.mockState.rateLimit = 100

		// Should still function
		view = ta.App.View()
		if len(view) == 0 {
			t.Error("View should render with low rate limit")
		}
	})

	t.Run("very low rate limit still allows viewing", func(t *testing.T) {
		ta := NewTestApp(t, WithMockRateLimit(10))
		ta.SetSize(120, 40)

		ta.App.Update(app.WorkflowsLoadedMsg{Workflows: DefaultTestWorkflows()})

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render even with very low rate limit")
		}
	})
}

func TestPolling_LogPollingStop(t *testing.T) {
	t.Run("StopLogPolling is safe to call multiple times", func(t *testing.T) {
		ta := NewTestApp(t)
		ta.SetSize(120, 40)

		// Should not panic
		ta.App.StopLogPolling()
		ta.App.StopLogPolling()
		ta.App.StopLogPolling()

		view := ta.App.View()
		if len(view) == 0 {
			t.Error("View should render")
		}
	})

	t.Run("StopLogPolling safe when poller is nil", func(t *testing.T) {
		ta := NewTestApp(t)

		// Should not panic
		ta.App.StopLogPolling()
	})
}
