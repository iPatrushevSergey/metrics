package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewZapLoggerAdapter(t *testing.T) {
	zl := zap.NewNop()
	adapter := NewZapLoggerAdapter(zl)
	require.NotNil(t, adapter)
	adapter.Debug("d")
	adapter.Info("i")
	adapter.Warn("w")
	adapter.Error("e")
}

func TestInitialize_valid(t *testing.T) {
	zl, err := Initialize("info")
	require.NoError(t, err)
	assert.NotNil(t, zl)
}

func TestInitialize_invalid(t *testing.T) {
	_, err := Initialize("invalid_level_xyz")
	require.Error(t, err)
}
