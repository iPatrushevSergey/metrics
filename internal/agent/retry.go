package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"

	"github.com/cenkalti/backoff/v5"
	"github.com/iPatrushevSergey/metrics/internal/hash"
	"github.com/iPatrushevSergey/metrics/internal/reqcrypto"
	"github.com/iPatrushevSergey/metrics/internal/retry"
)

// DefaultRetryConfig returns the default configuration: 3 repetitions with 1s, 3s, 5s intervals
func DefaultRetryConfig() retry.RetryConfig {
	return retry.DefaultRetryConfig()
}

// isHTTPStatusRetriable determines whether the HTTP request can be repeated for this error
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

// sendRequestWithRetry sends an HTTP request with retry logic using a fixed interval backoff strategy.
func sendRequestWithRetry(
	ctx context.Context,
	client *http.Client,
	config retry.RetryConfig,
	url string,
	bodyBytes []byte,
	key string,
	pub *rsa.PublicKey,
) (*http.Response, error) {
	wireBody := bodyBytes
	contentType := "application/json"
	if pub != nil {
		enc, err := reqcrypto.Seal(pub, bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("encrypt body: %w", err)
		}
		wireBody = enc
		contentType = "application/octet-stream"
	}

	// Compress body once before retry loop
	var compressedBuf bytes.Buffer
	gz := gzip.NewWriter(&compressedBuf)
	if _, err := gz.Write(wireBody); err != nil {
		return nil, fmt.Errorf("error compressing body: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("error closing gzip writer: %w", err)
	}
	compressedBody := compressedBuf.Bytes()

	backoffStrategy := retry.NewFixedIntervalBackoff(config)

	response, err := backoff.Retry(ctx, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(compressedBody))
		if err != nil {
			return nil, backoff.Permanent(fmt.Errorf("request creation error: %w", err))
		}
		req.Header.Add("Content-Type", contentType)
		req.Header.Add("Content-Encoding", "gzip")
		req.Header.Add("Accept-Encoding", "gzip")
		if pub != nil {
			req.Header.Set(reqcrypto.HeaderName, reqcrypto.HeaderValue)
		}

		if key != "" {
			hashValue := hash.CalculateHash(bodyBytes, key)
			req.Header.Add("HashSHA256", hashValue)
		}

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
