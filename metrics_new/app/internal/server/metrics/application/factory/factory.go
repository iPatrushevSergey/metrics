package factory

import (
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/usecase"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/service"
)

// MetricUseCasesParams contains dependencies required to build metrics use cases.
type MetricUseCasesParams struct {
	MetricRepo     port.MetricRepository
	MetricSvc      service.MetricService
	Transactor     port.Transactor
	MetricFileRepo port.MetricFileRepository
	SyncFileWrites bool
	AuditFileRepo  port.AuditFileRepository
	AuditGateway   port.AuditGateway
	AuditPublisher port.AuditPublisher
}

// MetricUseCases holds metrics module use cases exposed to the composition root.
type MetricUseCases struct {
	GetMetricValue         port.UseCase[dto.GetMetricValueInput, string]
	GetMetric              port.UseCase[dto.GetMetricInput, dto.MetricOutput]
	UpdateMetric           port.UseCase[dto.UpdateMetricInput, struct{}]
	UpsertMetric           port.UseCase[dto.UpsertMetricInput, struct{}]
	UpsertMetricsBatch     port.UseCase[dto.UpsertMetricsBatchInput, struct{}]
	GetAllMetrics          port.UseCase[struct{}, []dto.MetricForDisplayOutput]
	PingDB                 port.UseCase[struct{}, struct{}]
	MetricsSnapshot        port.UseCase[struct{}, int]
	RestoreMetricsFromFile port.UseCase[struct{}, struct{}]
	RecordAuditToFile      port.UseCase[dto.AuditEvent, struct{}]
	CreateAudit            port.UseCase[dto.AuditEvent, struct{}]
}

// NewMetricUseCases builds metrics module use cases.
func NewMetricUseCases(p MetricUseCasesParams) *MetricUseCases {
	var metricFileRepo port.MetricFileRepository
	if p.MetricFileRepo != nil && p.SyncFileWrites {
		metricFileRepo = p.MetricFileRepo
	}

	var (
		metricsSnapshot        port.UseCase[struct{}, int]
		restoreMetricsFromFile port.UseCase[struct{}, struct{}]
		recordAuditToFile      port.UseCase[dto.AuditEvent, struct{}]
		createAudit            port.UseCase[dto.AuditEvent, struct{}]
	)
	if p.MetricFileRepo != nil {
		metricsSnapshot = usecase.NewMetricsSnapshot(p.MetricRepo, p.MetricFileRepo)
		restoreMetricsFromFile = usecase.NewRestoreMetricsFromFile(p.MetricRepo, p.MetricFileRepo)
	}
	if p.AuditFileRepo != nil {
		recordAuditToFile = usecase.NewRecordAuditToFile(p.AuditFileRepo)
	}
	if p.AuditGateway != nil {
		createAudit = usecase.NewCreateAudit(p.AuditGateway)
	}

	return &MetricUseCases{
		GetMetricValue:         usecase.NewGetMetricValue(p.MetricRepo, p.MetricSvc),
		GetMetric:              usecase.NewGetMetric(p.MetricRepo, p.MetricSvc),
		UpdateMetric:           usecase.NewUpdateMetric(p.MetricRepo, p.MetricSvc, metricFileRepo, p.AuditPublisher),
		UpsertMetric:           usecase.NewUpsertMetric(p.MetricRepo, p.MetricSvc, metricFileRepo, p.AuditPublisher),
		UpsertMetricsBatch:     usecase.NewUpsertMetricsBatch(p.MetricRepo, p.MetricSvc, p.Transactor, metricFileRepo, p.AuditPublisher),
		GetAllMetrics:          usecase.NewGetAllMetrics(p.MetricRepo, p.MetricSvc),
		PingDB:                 usecase.NewPingDB(p.MetricRepo),
		MetricsSnapshot:        metricsSnapshot,
		RestoreMetricsFromFile: restoreMetricsFromFile,
		RecordAuditToFile:      recordAuditToFile,
		CreateAudit:            createAudit,
	}
}
