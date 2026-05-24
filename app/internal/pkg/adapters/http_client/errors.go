package httpclient

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
)

// RetriableStatusError wrapper over the status code.
type RetriableStatusError struct {
	Code int
}

// Error returns a string representation of the error.
func (e *RetriableStatusError) Error() string {
	return fmt.Sprintf("retriable HTTP status: %d", e.Code)
}

// IsRetriableHTTPStatus checks the repeatable HTTP status codes.
func IsRetriableHTTPStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests:
		return true
	default:
		return code >= http.StatusInternalServerError
	}
}

// IsRetriable checks whether HTTP/network error is repeatable.
func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	// Network-level errors (connection refused, timeout, broken pipe, EOF)
	if isNetworkError(err) {
		return true
	}

	// checking the status code of the error.
	var st *RetriableStatusError
	if errors.As(err, &st) {
		return IsRetriableHTTPStatus(st.Code)
	}

	return false
}

// isNetworkError checks for common transient network errors.
func isNetworkError(err error) bool {
	// EOF — connection was closed by the server
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	// Connection refused, reset, broken pipe
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) {
		return true
	}

	// OS-level timeout
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	// net.Error with Timeout (e.g. dial timeout, read timeout)
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}
