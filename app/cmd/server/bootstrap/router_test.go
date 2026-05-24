package bootstrap

import (
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	pkginmemory "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/inmemory"
	metricinmemory "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"

	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	f := NewUseCaseFactory(
		WithMetricRepo(metricinmemory.NewMetricMemoryRepository()),
		WithMetricSvc(service.MetricService{}),
		WithTransactor(pkginmemory.NewTransactor()),
	)

	r, err := NewRouter(f, logger.NewNopLogger(), "secret", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, r)
}
