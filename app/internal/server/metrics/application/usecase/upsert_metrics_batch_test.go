package usecase

import (
	"context"
	"testing"

	pkginmemory "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/inmemory"
	metricinmemory "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"

	"github.com/stretchr/testify/require"
)

func TestUpsertMetricsBatch_Execute_create(t *testing.T) {
	ctx := context.Background()
	repo := metricinmemory.NewMetricMemoryRepository()
	v := 1.0

	uc := NewUpsertMetricsBatch(
		repo,
		service.MetricService{},
		pkginmemory.NewTransactor(),
		nil,
		nil,
	)

	_, err := uc.Execute(ctx, dto.UpsertMetricsBatchInput{
		Metrics: []dto.UpsertMetricInput{
			{ID: "a", MType: "gauge", Value: &v},
		},
	})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "a")
	require.NoError(t, err)
	require.NotNil(t, got.Value)
}

func TestUpsertMetricsBatch_Execute_empty(t *testing.T) {
	uc := NewUpsertMetricsBatch(
		metricinmemory.NewMetricMemoryRepository(),
		service.MetricService{},
		pkginmemory.NewTransactor(),
		nil,
		nil,
	)
	_, err := uc.Execute(context.Background(), dto.UpsertMetricsBatchInput{})
	require.NoError(t, err)
}
