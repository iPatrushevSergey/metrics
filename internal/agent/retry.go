package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// RetryConfig configuration for the retry logic of HTTP requests
type RetryConfig struct {
	MaxRetries uint
	Intervals  []time.Duration
}

// DefaultRetryConfig returns the default configuration: 3 repetitions with 1s, 3s, 5s intervals
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		Intervals:  []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

// isHTTPErrorRetriable determines whether the HTTP request can be repeated for this error
func isHTTPStatusRetriable(statusCode int) bool {
	switch {
	case statusCode >= 500:
		// Server errors (5xx) - retriable
		return true
	case statusCode == 429:
		// Too Many Requests - can be repeated after a delay
		return true
	case statusCode >= 400:
		// Client errors (4xx, except 429) - non-retriable
		return false
	default:
		// 2xx, 3xx - success, do not repeat
		return false
	}
}

// fixedIntervalBackoff implements backoff.BackOff at fixed intervals for HTTP
type fixedIntervalBackoff struct {
	intervals  []time.Duration
	maxRetries uint
	attempt    uint
}

// NextBackOff returns the next wait duration for a retry attempt, or backoff.Stop if no more retries remain.
func (b *fixedIntervalBackoff) NextBackOff() time.Duration {
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
func (b *fixedIntervalBackoff) Reset() {
	b.attempt = 0
}

// sendRequestWithRetry sends an HTTP request with retry logic using a fixed interval backoff strategy.
func sendRequestWithRetry(
	ctx context.Context,
	client *http.Client,
	config RetryConfig,
	url string,
	bodyBytes []byte,
) (*http.Response, error) {
	backoffStrategy := &fixedIntervalBackoff{
		intervals:  config.Intervals,
		maxRetries: config.MaxRetries,
	}

	response, err := backoff.Retry(ctx, func() (*http.Response, error) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(bodyBytes); err != nil {
			return nil, backoff.Permanent(fmt.Errorf("error compressing body: %w", err))
		}
		if err := gz.Close(); err != nil {
			return nil, backoff.Permanent(fmt.Errorf("error closing gzip writer: %w", err))
		}

		// Request formation
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
		if err != nil {
			return nil, backoff.Permanent(fmt.Errorf("request creation error: %w", err))
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Content-Encoding", "gzip")
		req.Header.Add("Accept-Encoding", "gzip")

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		statusCode := resp.StatusCode

		if isHTTPStatusRetriable(statusCode) {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP %d", statusCode)
		}

		return resp, nil
	}, backoff.WithBackOff(backoffStrategy), backoff.WithMaxTries(config.MaxRetries+1))

	if err != nil {
		return nil, err
	}

	return response, nil
}
