package usecase

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRestoreMetricsFromFile_Execute(t *testing.T) {
	ctx := context.Background()
	v := 3.0
	metrics := []entity.Metric{{ID: "m", MType: entity.Gauge, Value: &v}}

	ctrl := gomock.NewController(t)
	fileRepo := mocks.NewMockMetricFileRepository(ctrl)
	fileRepo.EXPECT().LoadAll(ctx).Return(metrics, nil)

	repo := inmemory.NewMetricMemoryRepository()
	uc := NewRestoreMetricsFromFile(repo, fileRepo)

	_, err := uc.Execute(ctx, struct{}{})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "m")
	require.NoError(t, err)
	require.Equal(t, v, *got.Value)
}

func TestRestoreMetricsFromFile_Execute_emptyFile(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	fileRepo := mocks.NewMockMetricFileRepository(ctrl)
	fileRepo.EXPECT().LoadAll(ctx).Return(nil, nil)

	repo := inmemory.NewMetricMemoryRepository()
	uc := NewRestoreMetricsFromFile(repo, fileRepo)
	_, err := uc.Execute(ctx, struct{}{})
	require.NoError(t, err)
}
