package http_client

import (
	"errors"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return false }

func TestIsRetriableHTTPStatus(t *testing.T) {
	assert.True(t, IsRetriableHTTPStatus(http.StatusTooManyRequests))
	assert.True(t, IsRetriableHTTPStatus(http.StatusInternalServerError))
	assert.False(t, IsRetriableHTTPStatus(http.StatusBadRequest))
}

func TestIsRetriable(t *testing.T) {
	assert.False(t, IsRetriable(nil))
	assert.True(t, IsRetriable(io.EOF))
	assert.True(t, IsRetriable(&RetriableStatusError{Code: http.StatusServiceUnavailable}))

	var netErr net.Error = &timeoutError{}
	assert.True(t, IsRetriable(netErr))
}

func TestRetriableStatusError(t *testing.T) {
	err := &RetriableStatusError{Code: 503}
	assert.Contains(t, err.Error(), "503")
	var st *RetriableStatusError
	assert.True(t, errors.As(err, &st))
}
