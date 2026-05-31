package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRun_fixtureDir checks that generation succeeds for the internal/fixture package.
// go test runs the binary in this package's source directory, so a relative path is enough.
func TestRun_fixtureDir(t *testing.T) {
	root := filepath.Join("internal", "fixture")
	require.NoError(t, run(root))
}
