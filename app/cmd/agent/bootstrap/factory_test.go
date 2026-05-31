package bootstrap

import (
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewUseCaseFactory(t *testing.T) {
	ctrl := gomock.NewController(t)

	f := NewUseCaseFactory(
		WithMetricsRepo(mocks.NewMockMetricsRepository(ctrl)),
		WithMetricsSampler(mocks.NewMockMetricsSampler(ctrl)),
		WithMetricsGateway(mocks.NewMockMetricsGateway(ctrl)),
		WithLogger(mocks.NewMockLogger(ctrl)),
		WithRandFloat(func() float64 { return 1 }),
		WithReportInterval(time.Second),
	)

	assert.NotNil(t, f.PollRuntimeTick())
	assert.NotNil(t, f.PollGopsutilTick())
	assert.NotNil(t, f.ReportBatchTick())
}

func TestNewUseCaseFactory_missingSamplerPanics(t *testing.T) {
	ctrl := gomock.NewController(t)
	assert.Panics(t, func() {
		NewUseCaseFactory(
			WithMetricsRepo(mocks.NewMockMetricsRepository(ctrl)),
			WithMetricsGateway(mocks.NewMockMetricsGateway(ctrl)),
			WithLogger(mocks.NewMockLogger(ctrl)),
			WithRandFloat(func() float64 { return 1 }),
			WithReportInterval(time.Second),
		)
	})
}

func TestNewUseCaseFactory_missingGatewayPanics(t *testing.T) {
	ctrl := gomock.NewController(t)
	assert.Panics(t, func() {
		NewUseCaseFactory(
			WithMetricsRepo(mocks.NewMockMetricsRepository(ctrl)),
			WithMetricsSampler(mocks.NewMockMetricsSampler(ctrl)),
			WithLogger(mocks.NewMockLogger(ctrl)),
			WithRandFloat(func() float64 { return 1 }),
			WithReportInterval(time.Second),
		)
	})
}
