package usecase

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/audit"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/metrics"
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

func TestUpsertMetric_Execute_update(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMetricMemoryRepository()
	v1, v2 := 1.0, 2.0

	uc := NewUpsertMetric(repo, service.MetricService{}, nil, nil)
	_, err := uc.Execute(ctx, dto.UpsertMetricInput{ID: "x", MType: "gauge", Value: &v1})
	require.NoError(t, err)
	_, err = uc.Execute(ctx, dto.UpsertMetricInput{ID: "x", MType: "gauge", Value: &v2})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "x")
	require.NoError(t, err)
	assert.Equal(t, v2, *got.Value)
}

func TestUpsertMetric_Execute_badValue(t *testing.T) {
	repo := inmemory.NewMetricMemoryRepository()
	uc := NewUpsertMetric(repo, service.MetricService{}, nil, nil)
	_, err := uc.Execute(context.Background(), dto.UpsertMetricInput{
		ID: "x", MType: "gauge", Value: nil,
	})
	assert.ErrorIs(t, err, application.ErrBadMetricValue)
}

func TestUpsertMetric_Execute_withAudit(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMetricMemoryRepository()
	pub := audit.NewAuditEventPublisher(logger.NewNopLogger(), 10)
	v := 1.0

	uc := NewUpsertMetric(repo, service.MetricService{}, nil, pub)
	_, err := uc.Execute(ctx, dto.UpsertMetricInput{
		ID: "m", MType: "gauge", Value: &v, IPAddress: "127.0.0.1",
	})
	require.NoError(t, err)
}

func TestUpsertMetric_Execute_counter(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMetricMemoryRepository()
	d := int64(10)

	uc := NewUpsertMetric(repo, service.MetricService{}, nil, nil)
	_, err := uc.Execute(ctx, dto.UpsertMetricInput{ID: "hits", MType: "counter", Delta: &d})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "hits")
	require.NoError(t, err)
	assert.Equal(t, int64(10), *got.Delta)
}

func TestUpsertMetric_Execute_withFileSnapshot(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMetricMemoryRepository()
	fileRepo := metrics.NewMetricFileRepository(t.TempDir() + "/m.json")
	v := 2.0

	uc := NewUpsertMetric(repo, service.MetricService{}, fileRepo, nil)
	_, err := uc.Execute(ctx, dto.UpsertMetricInput{ID: "f", MType: "gauge", Value: &v})
	require.NoError(t, err)

	all, err := fileRepo.LoadAll(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)
}
