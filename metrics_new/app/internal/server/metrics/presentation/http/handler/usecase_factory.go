package handler

import (
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
)

// UseCaseFactory provides access to the metrics server use cases.
type UseCaseFactory interface {
	GetMetricValueUseCase() port.UseCase[dto.GetMetricValueInput, string]
	GetMetricUseCase() port.UseCase[dto.GetMetricInput, dto.MetricOutput]
	UpdateMetricUseCase() port.UseCase[dto.UpdateMetricInput, struct{}]
	UpsertMetricUseCase() port.UseCase[dto.UpsertMetricInput, struct{}]
	UpsertMetricsBatchUseCase() port.UseCase[dto.UpsertMetricsBatchInput, struct{}]
	GetAllMetricsUseCase() port.UseCase[struct{}, []dto.MetricForDisplayOutput]
	PingDBUseCase() port.UseCase[struct{}, struct{}]
}
