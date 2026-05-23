package bootstrap

import (
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	pkginmemory "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/audit"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/metrics"
	metricinmemory "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"

	"github.com/stretchr/testify/assert"
)

func TestNewUseCaseFactory(t *testing.T) {
	dir := t.TempDir()
	fileRepo := metrics.NewMetricFileRepository(dir + "/m.json")
	pub := audit.NewAuditEventPublisher(logger.NewNopLogger(), 4)

	f := NewUseCaseFactory(
		WithMetricRepo(metricinmemory.NewMetricMemoryRepository()),
		WithMetricSvc(service.MetricService{}),
		WithTransactor(pkginmemory.NewTransactor()),
		WithMetricFileRepo(fileRepo),
		WithSyncFileWrites(true),
		WithAuditPublisher(pub),
	)

	assert.NotNil(t, f.GetMetricValueUseCase())
	assert.NotNil(t, f.GetMetricUseCase())
	assert.NotNil(t, f.UpdateMetricUseCase())
	assert.NotNil(t, f.UpsertMetricUseCase())
	assert.NotNil(t, f.UpsertMetricsBatchUseCase())
	assert.NotNil(t, f.GetAllMetricsUseCase())
	assert.NotNil(t, f.PingDBUseCase())
	assert.NotNil(t, f.MetricsSnapshotUseCase())
	assert.NotNil(t, f.RestoreMetricsFromFileUseCase())
}

func TestNewUseCaseFactory_missingRepoPanics(t *testing.T) {
	assert.Panics(t, func() {
		NewUseCaseFactory(WithTransactor(pkginmemory.NewTransactor()))
	})
}

func TestWithOptions_setDeps(t *testing.T) {
	repo := metricinmemory.NewMetricMemoryRepository()
	var p factoryParams
	WithMetricRepo(repo)(&p)
	WithMetricSvc(service.MetricService{})(&p)
	WithTransactor(pkginmemory.NewTransactor())(&p)
	WithSyncFileWrites(true)(&p)

	assert.Equal(t, repo, p.metricRepo)
	assert.True(t, p.syncFileWrites)
}
