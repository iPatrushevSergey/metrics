package usecase

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertMetric_Execute(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMetricMemoryRepository()
	v := 2.0

	uc := NewUpsertMetric(repo, service.MetricService{}, nil, nil)
	_, err := uc.Execute(ctx, dto.UpsertMetricInput{ID: "mem", MType: "gauge", Value: &v})
	require.NoError(t, err)

	_, err = uc.Execute(ctx, dto.UpsertMetricInput{ID: "mem", MType: "bad"})
	assert.ErrorIs(t, err, application.ErrBadMetricType)
}
