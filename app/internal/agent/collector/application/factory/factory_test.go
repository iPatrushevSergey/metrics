package factory

import (
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewUseCases(t *testing.T) {
	ctrl := gomock.NewController(t)

	uc := NewUseCases(Params{
		MetricsRepo:    mocks.NewMockMetricsRepository(ctrl),
		MetricsSampler: mocks.NewMockMetricsSampler(ctrl),
		MetricsGateway: mocks.NewMockMetricsGateway(ctrl),
		Log:            mocks.NewMockLogger(ctrl),
		RandFloat:      func() float64 { return 1 },
		ReportInterval: time.Second,
	})

	assert.NotNil(t, uc.PollRuntime)
	assert.NotNil(t, uc.PollGopsutil)
	assert.NotNil(t, uc.ReportBatch)
}
