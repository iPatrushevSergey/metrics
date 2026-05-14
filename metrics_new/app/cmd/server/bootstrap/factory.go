package bootstrap

import (
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/option"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/factory"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/service"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/presentation/http/handler"
)

// UseCaseFactory provides all module use cases needed by the composition root.
type UseCaseFactory interface {
	handler.UseCaseFactory
}

// useCaseFactory implements UseCaseFactory.
type useCaseFactory struct {
	getMetricValue     port.UseCase[dto.GetMetricValueInput, string]
	getMetric          port.UseCase[dto.GetMetricInput, dto.MetricOutput]
	updateMetric       port.UseCase[dto.UpdateMetricInput, struct{}]
	upsertMetric       port.UseCase[dto.UpsertMetricInput, struct{}]
	upsertMetricsBatch port.UseCase[dto.UpsertMetricsBatchInput, struct{}]
	getAllMetrics      port.UseCase[struct{}, []dto.MetricForDisplayOutput]
	pingDB             port.UseCase[struct{}, struct{}]
}

var _ handler.UseCaseFactory = (*useCaseFactory)(nil)

// NewUseCaseFactory builds the use case factory using functional options.
func NewUseCaseFactory(opts ...option.Option[factoryParams]) UseCaseFactory {
	p := factoryParams{}
	option.Apply(&p, opts...)
	p.validate()

	metricUseCases := buildMetricUseCases(p)
	return &useCaseFactory{
		getMetricValue:     metricUseCases.GetMetricValue,
		getMetric:          metricUseCases.GetMetric,
		updateMetric:       metricUseCases.UpdateMetric,
		upsertMetric:       metricUseCases.UpsertMetric,
		upsertMetricsBatch: metricUseCases.UpsertMetricsBatch,
		getAllMetrics:      metricUseCases.GetAllMetrics,
		pingDB:             metricUseCases.PingDB,
	}
}

// GetMetricValueUseCase returns the get metric value use case.
func (f *useCaseFactory) GetMetricValueUseCase() port.UseCase[dto.GetMetricValueInput, string] {
	return f.getMetricValue
}

// GetMetricUseCase returns the get metric use case.
func (f *useCaseFactory) GetMetricUseCase() port.UseCase[dto.GetMetricInput, dto.MetricOutput] {
	return f.getMetric
}

// UpdateMetricUseCase returns the update metric use case.
func (f *useCaseFactory) UpdateMetricUseCase() port.UseCase[dto.UpdateMetricInput, struct{}] {
	return f.updateMetric
}

// UpsertMetricUseCase returns the upsert metric use case.
func (f *useCaseFactory) UpsertMetricUseCase() port.UseCase[dto.UpsertMetricInput, struct{}] {
	return f.upsertMetric
}

// UpsertMetricsBatchUseCase returns the upsert metrics batch use case.
func (f *useCaseFactory) UpsertMetricsBatchUseCase() port.UseCase[dto.UpsertMetricsBatchInput, struct{}] {
	return f.upsertMetricsBatch
}

// GetAllMetricsUseCase returns the get all metrics use case.
func (f *useCaseFactory) GetAllMetricsUseCase() port.UseCase[struct{}, []dto.MetricForDisplayOutput] {
	return f.getAllMetrics
}

// PingDBUseCase returns the ping db use case.
func (f *useCaseFactory) PingDBUseCase() port.UseCase[struct{}, struct{}] {
	return f.pingDB
}

// factoryParams holds all dependencies needed to build the use case factory.
type factoryParams struct {
	metricRepo port.MetricRepository
	metricSvc  service.MetricService
}

// validate checks if all required dependencies are set.
func (p factoryParams) validate() {
	if p.metricRepo == nil {
		panic("NewUseCaseFactory: WithMetricRepo is required")
	}
}

// metricUseCasesParams builds the metric use cases parameters.
func (p factoryParams) metricUseCasesParams() factory.MetricUseCasesParams {
	return factory.MetricUseCasesParams{
		MetricRepo: p.metricRepo,
		MetricSvc:  p.metricSvc,
	}
}

// buildMetricUseCases builds the metric use cases from the factory parameters.
func buildMetricUseCases(p factoryParams) *factory.MetricUseCases {
	return factory.NewMetricUseCases(p.metricUseCasesParams())
}

// WithMetricRepo sets the metric repository dependency.
func WithMetricRepo(r port.MetricRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.metricRepo = r }
}

// WithMetricSvc sets the metric service dependency.
func WithMetricSvc(s service.MetricService) option.Option[factoryParams] {
	return func(p *factoryParams) { p.metricSvc = s }
}
