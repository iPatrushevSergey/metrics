package worker

import (
	"context"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	presport "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/port"
)

type stubUseCase[In, Out any] struct {
	onExecute func(context.Context, In) (Out, error)
}

func (s stubUseCase[In, Out]) Execute(ctx context.Context, in In) (Out, error) {
	if s.onExecute != nil {
		return s.onExecute(ctx, in)
	}
	var zero Out
	return zero, nil
}

type stubUseCaseFactory struct {
	metricsSnapshot port.UseCase[struct{}, int]
	recordAudit     port.UseCase[dto.AuditEvent, struct{}]
	createRemote    port.UseCase[dto.AuditEvent, struct{}]
}

func (f stubUseCaseFactory) GetMetricValueUseCase() port.UseCase[dto.GetMetricValueInput, string] {
	return nil
}
func (f stubUseCaseFactory) GetMetricUseCase() port.UseCase[dto.GetMetricInput, dto.MetricOutput] {
	return nil
}
func (f stubUseCaseFactory) UpdateMetricUseCase() port.UseCase[dto.UpdateMetricInput, struct{}] {
	return nil
}
func (f stubUseCaseFactory) UpsertMetricUseCase() port.UseCase[dto.UpsertMetricInput, struct{}] {
	return nil
}
func (f stubUseCaseFactory) UpsertMetricsBatchUseCase() port.UseCase[dto.UpsertMetricsBatchInput, struct{}] {
	return nil
}
func (f stubUseCaseFactory) GetAllMetricsUseCase() port.UseCase[struct{}, []dto.MetricForDisplayOutput] {
	return nil
}
func (f stubUseCaseFactory) PingDBUseCase() port.UseCase[struct{}, struct{}] { return nil }
func (f stubUseCaseFactory) MetricsSnapshotUseCase() port.UseCase[struct{}, int] {
	return f.metricsSnapshot
}
func (f stubUseCaseFactory) RestoreMetricsFromFileUseCase() port.UseCase[struct{}, struct{}] {
	return nil
}
func (f stubUseCaseFactory) RecordAuditToFileUseCase() port.UseCase[dto.AuditEvent, struct{}] {
	return f.recordAudit
}
func (f stubUseCaseFactory) CreateRemoteAuditUseCase() port.UseCase[dto.AuditEvent, struct{}] {
	return f.createRemote
}

var _ presport.UseCaseFactory = stubUseCaseFactory{}
