// Package retry provides configurable retry and backoff for transient failures.
package retry

import (
	"time"

	"github.com/cenkalti/backoff/v5"
)

// RetryConfig configuration for retry logic
type RetryConfig struct {
	MaxRetries uint
	Intervals  []time.Duration
}

// DefaultRetryConfig returns the default configuration: 3 repetitions with intervals of 1s, 3s, 5s
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		Intervals:  []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

// FixedIntervalBackoff implements backoff.BackOff with fixed intervals
type FixedIntervalBackoff struct {
	intervals  []time.Duration
	maxRetries uint
	attempt    uint
}

// NewFixedIntervalBackoff creates a new FixedIntervalBackoff with the given configuration
func NewFixedIntervalBackoff(config RetryConfig) *FixedIntervalBackoff {
	return &FixedIntervalBackoff{
		intervals:  config.Intervals,
		maxRetries: config.MaxRetries,
		attempt:    0,
	}
}

// NextBackOff returns the next wait duration for a retry attempt, or backoff.Stop if no more retries remain.
func (b *FixedIntervalBackoff) NextBackOff() time.Duration {
	if b.attempt >= b.maxRetries {
		return backoff.Stop
	}

	idx := int(b.attempt)
	if idx >= len(b.intervals) {
		idx = len(b.intervals) - 1
	}

	interval := b.intervals[idx]
	b.attempt++
	return interval
}

// Reset resets the number of retry attempts in the backoff strategy to zero.
func (b *FixedIntervalBackoff) Reset() {
	b.attempt = 0
}
