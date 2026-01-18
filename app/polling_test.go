package app

import (
	"testing"
	"time"
)

// TestNewAdaptivePoller tests the creation of a new AdaptivePoller.
func TestNewAdaptivePoller(t *testing.T) {
	getRateLimit := func() int { return 5000 }

	poller := NewAdaptivePoller(getRateLimit)

	if poller == nil {
		t.Fatal("NewAdaptivePoller returned nil")
	}
	if poller.baseInterval != 2*time.Second {
		t.Errorf("expected baseInterval 2s, got %v", poller.baseInterval)
	}
	if poller.maxInterval != 30*time.Second {
		t.Errorf("expected maxInterval 30s, got %v", poller.maxInterval)
	}
	if poller.getRateLimit == nil {
		t.Error("expected getRateLimit to be non-nil")
	}
}

// TestAdaptivePoller_NextInterval_DefaultRemaining tests NextInterval with high remaining count.
func TestAdaptivePoller_NextInterval_DefaultRemaining(t *testing.T) {
	// remaining >= 1000: return baseInterval (2s)
	getRateLimit := func() int { return 5000 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 2 * time.Second
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_NextInterval_Remaining1000 tests NextInterval at exactly 1000.
func TestAdaptivePoller_NextInterval_Remaining1000(t *testing.T) {
	// remaining == 1000: return baseInterval (2s)
	getRateLimit := func() int { return 1000 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 2 * time.Second
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_NextInterval_LessThan1000 tests NextInterval with remaining < 1000.
func TestAdaptivePoller_NextInterval_LessThan1000(t *testing.T) {
	// remaining < 1000 && remaining >= 500: return baseInterval * 1.5 (3s)
	testCases := []int{999, 800, 600, 500}
	expected := 3 * time.Second

	for _, remaining := range testCases {
		t.Run("remaining_"+string(rune(remaining)), func(t *testing.T) {
			getRateLimit := func() int { return remaining }
			poller := NewAdaptivePoller(getRateLimit)

			interval := poller.NextInterval()

			if interval != expected {
				t.Errorf("for remaining=%d, expected %v, got %v", remaining, expected, interval)
			}
		})
	}
}

// TestAdaptivePoller_NextInterval_LessThan500 tests NextInterval with remaining < 500.
func TestAdaptivePoller_NextInterval_LessThan500(t *testing.T) {
	// remaining < 500 && remaining >= 100: return baseInterval * 2 (4s)
	testCases := []int{499, 300, 200, 100}
	expected := 4 * time.Second

	for _, remaining := range testCases {
		t.Run("remaining_"+string(rune(remaining)), func(t *testing.T) {
			getRateLimit := func() int { return remaining }
			poller := NewAdaptivePoller(getRateLimit)

			interval := poller.NextInterval()

			if interval != expected {
				t.Errorf("for remaining=%d, expected %v, got %v", remaining, expected, interval)
			}
		})
	}
}

// TestAdaptivePoller_NextInterval_LessThan100 tests NextInterval with remaining < 100.
func TestAdaptivePoller_NextInterval_LessThan100(t *testing.T) {
	// remaining < 100: return maxInterval (30s)
	testCases := []int{99, 50, 10, 1, 0}
	expected := 30 * time.Second

	for _, remaining := range testCases {
		t.Run("remaining_"+string(rune(remaining)), func(t *testing.T) {
			getRateLimit := func() int { return remaining }
			poller := NewAdaptivePoller(getRateLimit)

			interval := poller.NextInterval()

			if interval != expected {
				t.Errorf("for remaining=%d, expected %v, got %v", remaining, expected, interval)
			}
		})
	}
}

// TestAdaptivePoller_NextInterval_Boundary99 tests NextInterval at boundary 99.
func TestAdaptivePoller_NextInterval_Boundary99(t *testing.T) {
	getRateLimit := func() int { return 99 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 30 * time.Second
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_NextInterval_Boundary100 tests NextInterval at boundary 100.
func TestAdaptivePoller_NextInterval_Boundary100(t *testing.T) {
	getRateLimit := func() int { return 100 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 4 * time.Second // baseInterval * 2
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_NextInterval_Boundary499 tests NextInterval at boundary 499.
func TestAdaptivePoller_NextInterval_Boundary499(t *testing.T) {
	getRateLimit := func() int { return 499 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 4 * time.Second // baseInterval * 2
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_NextInterval_Boundary500 tests NextInterval at boundary 500.
func TestAdaptivePoller_NextInterval_Boundary500(t *testing.T) {
	getRateLimit := func() int { return 500 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 3 * time.Second // baseInterval * 1.5
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_NextInterval_Boundary999 tests NextInterval at boundary 999.
func TestAdaptivePoller_NextInterval_Boundary999(t *testing.T) {
	getRateLimit := func() int { return 999 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 3 * time.Second // baseInterval * 1.5
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_NextInterval_NegativeRemaining tests NextInterval with negative remaining.
func TestAdaptivePoller_NextInterval_NegativeRemaining(t *testing.T) {
	// Negative remaining should be treated as very low
	getRateLimit := func() int { return -1 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 30 * time.Second
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_NextInterval_ZeroRemaining tests NextInterval with zero remaining.
func TestAdaptivePoller_NextInterval_ZeroRemaining(t *testing.T) {
	getRateLimit := func() int { return 0 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 30 * time.Second
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_NextInterval_VeryHighRemaining tests NextInterval with very high remaining.
func TestAdaptivePoller_NextInterval_VeryHighRemaining(t *testing.T) {
	getRateLimit := func() int { return 100000 }
	poller := NewAdaptivePoller(getRateLimit)

	interval := poller.NextInterval()

	expected := 2 * time.Second
	if interval != expected {
		t.Errorf("expected %v, got %v", expected, interval)
	}
}

// TestAdaptivePoller_DynamicRateLimit tests that NextInterval uses the current rate limit.
func TestAdaptivePoller_DynamicRateLimit(t *testing.T) {
	remaining := 5000
	getRateLimit := func() int { return remaining }
	poller := NewAdaptivePoller(getRateLimit)

	// Initially high remaining
	interval1 := poller.NextInterval()
	if interval1 != 2*time.Second {
		t.Errorf("expected 2s, got %v", interval1)
	}

	// Simulate rate limit decreasing
	remaining = 50
	interval2 := poller.NextInterval()
	if interval2 != 30*time.Second {
		t.Errorf("expected 30s, got %v", interval2)
	}

	// Rate limit recovers
	remaining = 1000
	interval3 := poller.NextInterval()
	if interval3 != 2*time.Second {
		t.Errorf("expected 2s, got %v", interval3)
	}
}

// TestAdaptivePoller_ConsecutiveCalls tests multiple consecutive NextInterval calls.
func TestAdaptivePoller_ConsecutiveCalls(t *testing.T) {
	getRateLimit := func() int { return 5000 }
	poller := NewAdaptivePoller(getRateLimit)

	for i := 0; i < 100; i++ {
		interval := poller.NextInterval()
		if interval != 2*time.Second {
			t.Errorf("iteration %d: expected 2s, got %v", i, interval)
		}
	}
}

// TestAdaptivePoller_IntervalValues tests that all interval values are correct.
func TestAdaptivePoller_IntervalValues(t *testing.T) {
	tests := []struct {
		name      string
		remaining int
		expected  time.Duration
	}{
		{"high_remaining", 5000, 2 * time.Second},
		{"exactly_1000", 1000, 2 * time.Second},
		{"999_light_caution", 999, 3 * time.Second},
		{"700_light_caution", 700, 3 * time.Second},
		{"500_light_caution", 500, 3 * time.Second},
		{"499_caution", 499, 4 * time.Second},
		{"300_caution", 300, 4 * time.Second},
		{"100_caution", 100, 4 * time.Second},
		{"99_critical", 99, 30 * time.Second},
		{"50_critical", 50, 30 * time.Second},
		{"1_critical", 1, 30 * time.Second},
		{"0_critical", 0, 30 * time.Second},
		{"negative_critical", -10, 30 * time.Second},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getRateLimit := func() int { return tc.remaining }
			poller := NewAdaptivePoller(getRateLimit)

			interval := poller.NextInterval()

			if interval != tc.expected {
				t.Errorf("for remaining=%d, expected %v, got %v", tc.remaining, tc.expected, interval)
			}
		})
	}
}
