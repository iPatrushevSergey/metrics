package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZapLogger(t *testing.T) {
	log, err := NewZapLogger(Config{Level: "info"})
	require.NoError(t, err)
	log.Info("test")
	assert.NoError(t, log.Sync())
}

func TestNewZapLogger_invalidLevel(t *testing.T) {
	_, err := NewZapLogger(Config{Level: "not-a-level"})
	assert.Error(t, err)
}
