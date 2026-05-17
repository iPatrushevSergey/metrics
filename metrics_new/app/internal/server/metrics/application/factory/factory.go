package factory

import (
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/usecase"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/service"
)

// MetricUseCaseParams contains dependencies required to build metrics use cases.
type MetricUseCasesParams struct {
	MetricRepo port.MetricRepository
	MetricSvc  service.MetricService
	Transactor port.Transactor
}

// MetricUseCases holds metrics module use cases exposed to the composition root.
type MetricUseCases struct {
	GetMetricValue     port.UseCase[dto.GetMetricValueInput, string]
	GetMetric          port.UseCase[dto.GetMetricInput, dto.MetricOutput]
	UpdateMetric       port.UseCase[dto.UpdateMetricInput, struct{}]
	UpsertMetric       port.UseCase[dto.UpsertMetricInput, struct{}]
	UpsertMetricsBatch port.UseCase[dto.UpsertMetricsBatchInput, struct{}]
	GetAllMetrics      port.UseCase[struct{}, []dto.MetricForDisplayOutput]
	PingDB             port.UseCase[struct{}, struct{}]
}

// NewMetricUseCases builds metrics module use cases.
func NewMetricUseCases(p MetricUseCasesParams) *MetricUseCases {
	return &MetricUseCases{
		GetMetricValue:     usecase.NewGetMetricValue(p.MetricRepo, p.MetricSvc),
		GetMetric:          usecase.NewGetMetric(p.MetricRepo, p.MetricSvc),
		UpdateMetric:       usecase.NewUpdateMetric(p.MetricRepo, p.MetricSvc),
		UpsertMetric:       usecase.NewUpsertMetric(p.MetricRepo, p.MetricSvc),
		UpsertMetricsBatch: usecase.NewUpsertMetricsBatch(p.MetricRepo, p.MetricSvc, p.Transactor),
		GetAllMetrics:      usecase.NewGetAllMetrics(p.MetricRepo, p.MetricSvc),
		PingDB:             usecase.NewPingDB(p.MetricRepo),
	}
}
