package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseListenAddress(t *testing.T) {
	addr, err := parseListenAddress("localhost:8080")
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", addr)
}

func TestParseDuration(t *testing.T) {
	d, err := parseDuration("30")
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, d)

	d, err = parseDuration(5)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, d)
}
