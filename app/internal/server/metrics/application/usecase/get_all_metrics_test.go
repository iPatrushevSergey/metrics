package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAllMetrics_Execute(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMetricMemoryRepository()
	v := 42.0
	require.NoError(t, repo.Create(ctx, entity.Metric{ID: "cpu", MType: entity.Gauge, Value: &v}))

	uc := NewGetAllMetrics(repo, service.MetricService{})
	out, err := uc.Execute(ctx, struct{}{})

	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "cpu", out[0].MetricID)
	assert.Contains(t, out[0].MetricValue, "42")
}

func TestGetAllMetrics_Execute_error(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockMetricReader(ctrl)
	reader.EXPECT().GetAll(ctx).Return(nil, errors.New("db"))

	uc := NewGetAllMetrics(reader, service.MetricService{})
	_, err := uc.Execute(ctx, struct{}{})
	assert.Error(t, err)
}
