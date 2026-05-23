package factory

import (
	"testing"

	pkginmemory "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/metrics"
	metricinmemory "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricUseCases(t *testing.T) {
	uc := NewMetricUseCases(MetricUseCasesParams{
		MetricRepo: metricinmemory.NewMetricMemoryRepository(),
		MetricSvc:  service.MetricService{},
		Transactor: pkginmemory.NewTransactor(),
	})

	assert.NotNil(t, uc.GetMetric)
	assert.NotNil(t, uc.UpsertMetricsBatch)
	assert.NotNil(t, uc.PingDB)
	assert.NotNil(t, uc.GetAllMetrics)
	assert.NotNil(t, uc.UpdateMetric)
}

func TestNewMetricUseCases_withFileRepo(t *testing.T) {
	fileRepo := metrics.NewMetricFileRepository(t.TempDir() + "/m.json")
	uc := NewMetricUseCases(MetricUseCasesParams{
		MetricRepo:     metricinmemory.NewMetricMemoryRepository(),
		MetricSvc:      service.MetricService{},
		Transactor:     pkginmemory.NewTransactor(),
		MetricFileRepo: fileRepo,
	})
	assert.NotNil(t, uc.MetricsSnapshot)
	assert.NotNil(t, uc.RestoreMetricsFromFile)
}
