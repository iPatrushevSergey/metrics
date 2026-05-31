// Package factory builds collector use cases.
package factory

import (
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/usecase"
)

// Params contains dependencies required to build collector use cases.
type Params struct {
	MetricsRepo    port.MetricsRepository
	MetricsSampler port.MetricsSampler
	MetricsGateway port.MetricsGateway
	Log            port.Logger
	RandFloat      func() float64
	ReportInterval time.Duration
}

// UseCases holds collector module use cases exposed to composition root.
type UseCases struct {
	PollRuntime  port.UseCase[struct{}, int]
	PollGopsutil port.UseCase[struct{}, int]
	ReportBatch  port.UseCase[struct{}, int]
}

// NewUseCases builds collector module use cases.
func NewUseCases(p Params) *UseCases {
	return &UseCases{
		PollRuntime:  usecase.NewPollRuntimeTick(p.MetricsSampler, p.MetricsRepo, p.RandFloat),
		PollGopsutil: usecase.NewPollGopsutilTick(p.MetricsSampler, p.MetricsRepo, p.Log),
		ReportBatch:  usecase.NewReportBatchTick(p.MetricsRepo, p.MetricsGateway, p.Log, p.ReportInterval),
	}
}
