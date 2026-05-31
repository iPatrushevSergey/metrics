package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricPostgresRepository_GetByIDs_empty(t *testing.T) {
	repo := NewMetricPostgresRepository(nil)
	got, err := repo.GetByIDs(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, got)
}
