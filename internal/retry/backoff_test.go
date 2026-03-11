package retry

import (
	"testing"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	assert.Equal(t, uint(3), cfg.MaxRetries)
	assert.Len(t, cfg.Intervals, 3)
	assert.Equal(t, 1*time.Second, cfg.Intervals[0])
	assert.Equal(t, 3*time.Second, cfg.Intervals[1])
	assert.Equal(t, 5*time.Second, cfg.Intervals[2])
}

func TestFixedIntervalBackoff(t *testing.T) {
	cfg := RetryConfig{MaxRetries: 2, Intervals: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}}
	b := NewFixedIntervalBackoff(cfg)

	assert.Equal(t, 10*time.Millisecond, b.NextBackOff())
	assert.Equal(t, 20*time.Millisecond, b.NextBackOff())
	assert.Equal(t, backoff.Stop, b.NextBackOff())

	b.Reset()
	assert.Equal(t, 10*time.Millisecond, b.NextBackOff())
}
