// Package testsupport provides shared helpers for contract, component, and e2e tests.
package testsupport

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/app/cmd/server/bootstrap"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	pkginmemory "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/inmemory"
	metricinmemory "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"
)

// StartMetricsServer runs the metrics HTTP API in memory.
func StartMetricsServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	f := bootstrap.NewUseCaseFactory(
		bootstrap.WithMetricRepo(metricinmemory.NewMetricMemoryRepository()),
		bootstrap.WithMetricSvc(service.MetricService{}),
		bootstrap.WithTransactor(pkginmemory.NewTransactor()),
	)
	r, err := bootstrap.NewRouter(f, logger.NewNopLogger(), "", nil, nil)
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}
