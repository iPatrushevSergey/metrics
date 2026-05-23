package handler

import (
	"context"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/factory"
)

type stubUseCase[In, Out any] struct {
	out Out
	err error
}

func (s stubUseCase[In, Out]) Execute(_ context.Context, _ In) (Out, error) {
	return s.out, s.err
}

type stubUseCaseFactory struct {
	getMetricValue    port.UseCase[dto.GetMetricValueInput, string]
	getMetric         port.UseCase[dto.GetMetricInput, dto.MetricOutput]
	updateMetric      port.UseCase[dto.UpdateMetricInput, struct{}]
	upsertMetric      port.UseCase[dto.UpsertMetricInput, struct{}]
	upsertMetricsBatch port.UseCase[dto.UpsertMetricsBatchInput, struct{}]
	getAllMetrics     port.UseCase[struct{}, []dto.MetricForDisplayOutput]
	pingDB            port.UseCase[struct{}, struct{}]
}

func (f stubUseCaseFactory) GetMetricValueUseCase() port.UseCase[dto.GetMetricValueInput, string] {
	return f.getMetricValue
}
func (f stubUseCaseFactory) GetMetricUseCase() port.UseCase[dto.GetMetricInput, dto.MetricOutput] {
	return f.getMetric
}
func (f stubUseCaseFactory) UpdateMetricUseCase() port.UseCase[dto.UpdateMetricInput, struct{}] {
	return f.updateMetric
}
func (f stubUseCaseFactory) UpsertMetricUseCase() port.UseCase[dto.UpsertMetricInput, struct{}] {
	return f.upsertMetric
}
func (f stubUseCaseFactory) UpsertMetricsBatchUseCase() port.UseCase[dto.UpsertMetricsBatchInput, struct{}] {
	return f.upsertMetricsBatch
}
func (f stubUseCaseFactory) GetAllMetricsUseCase() port.UseCase[struct{}, []dto.MetricForDisplayOutput] {
	return f.getAllMetrics
}
func (f stubUseCaseFactory) PingDBUseCase() port.UseCase[struct{}, struct{}] {
	return f.pingDB
}
func (f stubUseCaseFactory) MetricsSnapshotUseCase() port.UseCase[struct{}, int]       { return nil }
func (f stubUseCaseFactory) RestoreMetricsFromFileUseCase() port.UseCase[struct{}, struct{}] {
	return nil
}
func (f stubUseCaseFactory) RecordAuditToFileUseCase() port.UseCase[dto.AuditEvent, struct{}] {
	return nil
}
func (f stubUseCaseFactory) CreateRemoteAuditUseCase() port.UseCase[dto.AuditEvent, struct{}] {
	return nil
}

var _ factory.UseCaseFactory = stubUseCaseFactory{}
