package migrate

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationsMetricsDir(t *testing.T) {
	dir := MigrationsMetricsDir()
	_, err := os.Stat(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
}
