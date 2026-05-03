package bootstrap

import (
	"time"

	collectorfactory "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/factory"
	collectorport "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	collectorpresfactory "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/presentation/factory"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/option"
)

// UseCaseFactory provides all module use cases needed by composition root.
type UseCaseFactory interface {
	collectorpresfactory.UseCaseFactory
}

// useCaseFactory implements UseCaseFactory; built in composition root.
type useCaseFactory struct {
	pollRuntime  collectorport.BackgroundRunner
	pollGopsutil collectorport.BackgroundRunner
	reportBatch  collectorport.BackgroundRunner
}

// NewUseCaseFactory builds the use case factory using functional options.
func NewUseCaseFactory(opts ...option.Option[factoryParams]) UseCaseFactory {
	p := factoryParams{}
	option.Apply(&p, opts...)
	p.validate()

	ucs := buildCollectorUseCases(p)
	return &useCaseFactory{
		pollRuntime:  ucs.PollRuntime,
		pollGopsutil: ucs.PollGopsutil,
		reportBatch:  ucs.ReportBatch,
	}
}

// PollRuntimeTick returns the poll runtime tick use case.
func (f *useCaseFactory) PollRuntimeTick() collectorport.BackgroundRunner {
	return f.pollRuntime
}

// PollGopsutilTick returns the poll gopsutil tick use case.
func (f *useCaseFactory) PollGopsutilTick() collectorport.BackgroundRunner {
	return f.pollGopsutil
}

// ReportBatchTick returns the report batch tick use case.
func (f *useCaseFactory) ReportBatchTick() collectorport.BackgroundRunner {
	return f.reportBatch
}

// factoryParams holds all dependencies needed to build the use case factory.
type factoryParams struct {
	metricsRepo    collectorport.MetricsRepository
	metricsSampler collectorport.MetricsSampler
	metricsClient  collectorport.MetricsClient
	log            collectorport.Logger
	randFloat      func() float64
	reportInterval time.Duration
}

// validate checks if all required dependencies are set.
func (p factoryParams) validate() {
	if p.metricsRepo == nil {
		panic("NewUseCaseFactory: WithMetricsRepo is required")
	}
	if p.metricsSampler == nil {
		panic("NewUseCaseFactory: WithMetricsSampler is required")
	}
	if p.metricsClient == nil {
		panic("NewUseCaseFactory: WithMetricsClient is required")
	}
	if p.log == nil {
		panic("NewUseCaseFactory: WithLogger is required")
	}
	if p.randFloat == nil {
		panic("NewUseCaseFactory: WithRandFloat is required")
	}
	if p.reportInterval <= 0 {
		panic("NewUseCaseFactory: WithReportInterval must be positive")
	}
}

// collectorParams builds the collector factory parameters.
func (p factoryParams) collectorParams() collectorfactory.Params {
	return collectorfactory.Params{
		MetricsRepo:    p.metricsRepo,
		MetricsSampler: p.metricsSampler,
		MetricsClient:  p.metricsClient,
		Log:            p.log,
		RandFloat:      p.randFloat,
		ReportInterval: p.reportInterval,
	}
}

// buildCollectorUseCases builds the collector use cases from the factory parameters.
func buildCollectorUseCases(p factoryParams) *collectorfactory.UseCases {
	return collectorfactory.NewUseCases(p.collectorParams())
}

// WithMetricsRepo sets the metrics repository dependency.
func WithMetricsRepo(r collectorport.MetricsRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.metricsRepo = r }
}

// WithMetricsSampler sets the metrics sampler dependency.
func WithMetricsSampler(s collectorport.MetricsSampler) option.Option[factoryParams] {
	return func(p *factoryParams) { p.metricsSampler = s }
}

// WithMetricsClient sets the metrics client dependency.
func WithMetricsClient(c collectorport.MetricsClient) option.Option[factoryParams] {
	return func(p *factoryParams) { p.metricsClient = c }
}

// WithLogger sets the logger dependency.
func WithLogger(l collectorport.Logger) option.Option[factoryParams] {
	return func(p *factoryParams) { p.log = l }
}

// WithRandFloat sets the random float generator dependency.
func WithRandFloat(f func() float64) option.Option[factoryParams] {
	return func(p *factoryParams) { p.randFloat = f }
}

// WithReportInterval sets the report interval dependency.
func WithReportInterval(d time.Duration) option.Option[factoryParams] {
	return func(p *factoryParams) { p.reportInterval = d }
}
