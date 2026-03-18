// Package logger provides a shared zap logger and initialization by level.
package logger

import (
	"go.uber.org/zap"
)

// Log is the global logger used by packages that do not receive a Logger (e.g. filestorage). Defaults to no-op.
var Log *zap.Logger = zap.NewNop()

// Initialize builds a zap.Logger from the given level string and sets it as Log. Returns the same logger.
func Initialize(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	Log = zl
	return zl, nil
}
