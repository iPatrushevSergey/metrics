// Package port defines the presentation-layer contract for metrics use cases.
package port

import (
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
)

// UseCaseFactory provides metrics use cases to HTTP handlers and background workers.
type UseCaseFactory interface {
	// GetMetricValueUseCase reads a formatted metric value.
	GetMetricValueUseCase() port.UseCase[dto.GetMetricValueInput, string]
	// GetMetricUseCase reads one metric.
	GetMetricUseCase() port.UseCase[dto.GetMetricInput, dto.MetricOutput]
	// UpdateMetricUseCase updates one metric.
	UpdateMetricUseCase() port.UseCase[dto.UpdateMetricInput, struct{}]
	// UpsertMetricUseCase creates or updates one metric.
	UpsertMetricUseCase() port.UseCase[dto.UpsertMetricInput, struct{}]
	// UpsertMetricsBatchUseCase creates or updates a batch of metrics.
	UpsertMetricsBatchUseCase() port.UseCase[dto.UpsertMetricsBatchInput, struct{}]
	// GetAllMetricsUseCase lists all metrics.
	GetAllMetricsUseCase() port.UseCase[struct{}, []dto.MetricForDisplayOutput]
	// PingDBUseCase checks storage connectivity.
	PingDBUseCase() port.UseCase[struct{}, struct{}]
	// MetricsSnapshotUseCase persists in-memory metrics to a file snapshot.
	MetricsSnapshotUseCase() port.UseCase[struct{}, int]
	// RestoreMetricsFromFileUseCase loads metrics from a file on startup.
	RestoreMetricsFromFileUseCase() port.UseCase[struct{}, struct{}]
	// RecordAuditToFileUseCase appends one audit event to the audit log file.
	RecordAuditToFileUseCase() port.UseCase[dto.AuditEvent, struct{}]
	// CreateRemoteAuditUseCase sends one audit event to a remote endpoint.
	CreateRemoteAuditUseCase() port.UseCase[dto.AuditEvent, struct{}]
}
