package app

import (
	"time"
)

// AdaptivePoller adjusts polling intervals based on GitHub API rate limit remaining.
// When the rate limit is low, it increases the interval to avoid hitting the limit.
type AdaptivePoller struct {
	baseInterval time.Duration
	maxInterval  time.Duration
	getRateLimit func() int
}

// NewAdaptivePoller creates a new AdaptivePoller with the given rate limit getter function.
// The default base interval is 2 seconds and max interval is 30 seconds.
func NewAdaptivePoller(getRateLimit func() int) *AdaptivePoller {
	return &AdaptivePoller{
		baseInterval: 2 * time.Second,
		maxInterval:  30 * time.Second,
		getRateLimit: getRateLimit,
	}
}

// NextInterval calculates the next polling interval based on the current rate limit remaining.
// The logic is:
//   - remaining < 100:  return maxInterval (30s) - critical, minimize requests
//   - remaining < 500:  return baseInterval * 2 (4s) - caution level
//   - remaining < 1000: return baseInterval * 1.5 (3s) - light caution
//   - default:          return baseInterval (2s) - normal operation
func (p *AdaptivePoller) NextInterval() time.Duration {
	remaining := p.getRateLimit()

	switch {
	case remaining < 100:
		// Critical: remaining very low, use maximum interval
		return p.maxInterval
	case remaining < 500:
		// Caution: double the interval
		return p.baseInterval * 2
	case remaining < 1000:
		// Light caution: 1.5x the interval
		return time.Duration(float64(p.baseInterval) * 1.5)
	default:
		// Normal: use base interval
		return p.baseInterval
	}
}
