package app

import (
	"time"
)

// Rate limit threshold constants
const (
	// RateLimitCritical is the threshold below which polling uses maximum interval
	RateLimitCritical = 100
	// RateLimitCaution is the threshold below which polling interval is doubled
	RateLimitCaution = 500
	// RateLimitLight is the threshold below which polling interval is increased by 1.5x
	RateLimitLight = 1000
	// CautionMultiplier is the multiplier for the base interval at caution level
	CautionMultiplier = 2
	// LightCautionMultiplier is the multiplier for the base interval at light caution level
	LightCautionMultiplier = 1.5
	// DefaultBaseInterval is the default polling interval
	DefaultBaseInterval = 2 * time.Second
	// DefaultMaxInterval is the maximum polling interval when rate limited
	DefaultMaxInterval = 30 * time.Second
)

// AdaptivePoller adjusts polling intervals based on GitHub API rate limit remaining.
// When the rate limit is low, it increases the interval to avoid hitting the limit.
type AdaptivePoller struct {
	baseInterval time.Duration
	maxInterval  time.Duration
	getRateLimit func() int
}

// NewAdaptivePoller creates a new AdaptivePoller with the given rate limit getter function.
// The default base interval is DefaultBaseInterval and max interval is DefaultMaxInterval.
func NewAdaptivePoller(getRateLimit func() int) *AdaptivePoller {
	return &AdaptivePoller{
		baseInterval: DefaultBaseInterval,
		maxInterval:  DefaultMaxInterval,
		getRateLimit: getRateLimit,
	}
}

// NextInterval calculates the next polling interval based on the current rate limit remaining.
// The logic is:
//   - remaining < RateLimitCritical:  return maxInterval - critical, minimize requests
//   - remaining < RateLimitCaution:   return baseInterval * CautionMultiplier - caution level
//   - remaining < RateLimitLight:     return baseInterval * LightCautionMultiplier - light caution
//   - default:                        return baseInterval - normal operation
func (p *AdaptivePoller) NextInterval() time.Duration {
	remaining := p.getRateLimit()

	switch {
	case remaining < RateLimitCritical:
		// Critical: remaining very low, use maximum interval
		return p.maxInterval
	case remaining < RateLimitCaution:
		// Caution: double the interval
		return p.baseInterval * CautionMultiplier
	case remaining < RateLimitLight:
		// Light caution: 1.5x the interval
		return time.Duration(float64(p.baseInterval) * LightCautionMultiplier)
	default:
		// Normal: use base interval
		return p.baseInterval
	}
}
