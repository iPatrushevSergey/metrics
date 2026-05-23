package usecase

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetric_Execute_create(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMetricMemoryRepository()

	uc := NewUpdateMetric(repo, service.MetricService{}, nil, nil)
	_, err := uc.Execute(ctx, dto.UpdateMetricInput{
		MType: "gauge",
		ID:    "cpu",
		Value: "1.5",
	})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "cpu")
	require.NoError(t, err)
	assert.Equal(t, 1.5, *got.Value)
}
