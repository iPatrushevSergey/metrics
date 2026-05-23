package factory

import (
	"testing"

	pkginmemory "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/inmemory"
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
}
