package usecase

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	pkginmemory "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/audit"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/metrics"
	metricinmemory "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"

	"github.com/stretchr/testify/assert"
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

func TestUpsertMetricsBatch_Execute_badValue(t *testing.T) {
	uc := NewUpsertMetricsBatch(
		metricinmemory.NewMetricMemoryRepository(),
		service.MetricService{},
		pkginmemory.NewTransactor(),
		nil,
		nil,
	)
	_, err := uc.Execute(context.Background(), dto.UpsertMetricsBatchInput{
		Metrics: []dto.UpsertMetricInput{{ID: "a", MType: "gauge"}},
	})
	assert.ErrorIs(t, err, application.ErrBadMetricValue)
}

func TestUpsertMetricsBatch_Execute_badType(t *testing.T) {
	uc := NewUpsertMetricsBatch(
		metricinmemory.NewMetricMemoryRepository(),
		service.MetricService{},
		pkginmemory.NewTransactor(),
		nil,
		nil,
	)
	_, err := uc.Execute(context.Background(), dto.UpsertMetricsBatchInput{
		Metrics: []dto.UpsertMetricInput{{ID: "a", MType: "bad"}},
	})
	assert.ErrorIs(t, err, application.ErrBadMetricType)
}

func TestUpsertMetricsBatch_Execute_mergeDuplicates(t *testing.T) {
	ctx := context.Background()
	repo := metricinmemory.NewMetricMemoryRepository()
	v1, v2 := 1.0, 3.0

	uc := NewUpsertMetricsBatch(repo, service.MetricService{}, pkginmemory.NewTransactor(), nil, nil)
	_, err := uc.Execute(ctx, dto.UpsertMetricsBatchInput{
		Metrics: []dto.UpsertMetricInput{
			{ID: "a", MType: "gauge", Value: &v1},
			{ID: "a", MType: "gauge", Value: &v2},
		},
	})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "a")
	require.NoError(t, err)
	assert.Equal(t, v2, *got.Value)
}

func TestUpsertMetricsBatch_Execute_updateExisting(t *testing.T) {
	ctx := context.Background()
	repo := metricinmemory.NewMetricMemoryRepository()
	v1, v2 := 1.0, 5.0

	uc := NewUpsertMetricsBatch(repo, service.MetricService{}, pkginmemory.NewTransactor(), nil, nil)
	_, err := uc.Execute(ctx, dto.UpsertMetricsBatchInput{
		Metrics: []dto.UpsertMetricInput{{ID: "a", MType: "gauge", Value: &v1}},
	})
	require.NoError(t, err)

	_, err = uc.Execute(ctx, dto.UpsertMetricsBatchInput{
		Metrics: []dto.UpsertMetricInput{{ID: "a", MType: "gauge", Value: &v2}},
	})
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "a")
	require.NoError(t, err)
	assert.Equal(t, v2, *got.Value)
}

func TestUpsertMetricsBatch_Execute_withFileSnapshot(t *testing.T) {
	ctx := context.Background()
	repo := metricinmemory.NewMetricMemoryRepository()
	fileRepo := metrics.NewMetricFileRepository(t.TempDir() + "/m.json")
	v := 4.0

	uc := NewUpsertMetricsBatch(repo, service.MetricService{}, pkginmemory.NewTransactor(), fileRepo, nil)
	_, err := uc.Execute(ctx, dto.UpsertMetricsBatchInput{
		Metrics: []dto.UpsertMetricInput{{ID: "a", MType: "gauge", Value: &v}},
	})
	require.NoError(t, err)

	all, err := fileRepo.LoadAll(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)
}

func TestUpsertMetricsBatch_Execute_withAudit(t *testing.T) {
	ctx := context.Background()
	repo := metricinmemory.NewMetricMemoryRepository()
	pub := audit.NewAuditEventPublisher(logger.NewNopLogger(), 10)
	v := 1.0

	uc := NewUpsertMetricsBatch(repo, service.MetricService{}, pkginmemory.NewTransactor(), nil, pub)
	_, err := uc.Execute(ctx, dto.UpsertMetricsBatchInput{
		Metrics: []dto.UpsertMetricInput{{ID: "a", MType: "gauge", Value: &v}},
	})
	require.NoError(t, err)
}
