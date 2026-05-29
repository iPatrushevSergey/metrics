package bootstrap

import (
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/option"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/factory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"
	metricpresport "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/port"
)

// UseCaseFactory provides all module use cases needed by the composition root.
type UseCaseFactory interface {
	metricpresport.UseCaseFactory
}

// useCaseFactory implements UseCaseFactory.
type useCaseFactory struct {
	getMetricValue         port.UseCase[dto.GetMetricValueInput, string]
	getMetric              port.UseCase[dto.GetMetricInput, dto.MetricOutput]
	updateMetric           port.UseCase[dto.UpdateMetricInput, struct{}]
	upsertMetric           port.UseCase[dto.UpsertMetricInput, struct{}]
	upsertMetricsBatch     port.UseCase[dto.UpsertMetricsBatchInput, struct{}]
	getAllMetrics          port.UseCase[struct{}, []dto.MetricForDisplayOutput]
	pingDB                 port.UseCase[struct{}, struct{}]
	metricsSnapshot        port.UseCase[struct{}, int]
	restoreMetricsFromFile port.UseCase[struct{}, struct{}]
	recordAuditToFile      port.UseCase[dto.AuditEvent, struct{}]
	createRemoteAudit      port.UseCase[dto.AuditEvent, struct{}]
}

var _ metricpresport.UseCaseFactory = (*useCaseFactory)(nil)

// NewUseCaseFactory builds the use case factory using functional options.
func NewUseCaseFactory(opts ...option.Option[factoryParams]) UseCaseFactory {
	p := factoryParams{}
	option.Apply(&p, opts...)
	p.validate()

	metricUseCases := buildMetricUseCases(p)
	return &useCaseFactory{
		getMetricValue:         metricUseCases.GetMetricValue,
		getMetric:              metricUseCases.GetMetric,
		updateMetric:           metricUseCases.UpdateMetric,
		upsertMetric:           metricUseCases.UpsertMetric,
		upsertMetricsBatch:     metricUseCases.UpsertMetricsBatch,
		getAllMetrics:          metricUseCases.GetAllMetrics,
		pingDB:                 metricUseCases.PingDB,
		metricsSnapshot:        metricUseCases.MetricsSnapshot,
		restoreMetricsFromFile: metricUseCases.RestoreMetricsFromFile,
		recordAuditToFile:      metricUseCases.RecordAuditToFile,
		createRemoteAudit:      metricUseCases.CreateRemoteAudit,
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

// MetricsSnapshotUseCase returns the metrics snapshot use case.
func (f *useCaseFactory) MetricsSnapshotUseCase() port.UseCase[struct{}, int] {
	return f.metricsSnapshot
}

// RestoreMetricsFromFileUseCase returns the restore-from-file use case.
func (f *useCaseFactory) RestoreMetricsFromFileUseCase() port.UseCase[struct{}, struct{}] {
	return f.restoreMetricsFromFile
}

// RecordAuditToFileUseCase returns the record audit to file use case.
func (f *useCaseFactory) RecordAuditToFileUseCase() port.UseCase[dto.AuditEvent, struct{}] {
	return f.recordAuditToFile
}

// CreateRemoteAuditUseCase returns the create remote audit use case.
func (f *useCaseFactory) CreateRemoteAuditUseCase() port.UseCase[dto.AuditEvent, struct{}] {
	return f.createRemoteAudit
}

// factoryParams holds all dependencies needed to build the use case factory.
type factoryParams struct {
	metricRepo     port.MetricRepository
	metricSvc      service.MetricService
	transactor     port.Transactor
	metricFileRepo port.MetricFileRepository
	syncFileWrites bool
	auditFileRepo  port.AuditFileRepository
	auditGateway   port.AuditGateway
	auditPublisher port.AuditPublisher
}

// validate checks if all required dependencies are set.
func (p factoryParams) validate() {
	if p.metricRepo == nil {
		panic("NewUseCaseFactory: WithMetricRepo is required")
	}
	if p.transactor == nil {
		panic("NewUseCaseFactory: WithTransactor is required")
	}
}

// metricUseCasesParams builds the metric use cases parameters.
func (p factoryParams) metricUseCasesParams() factory.MetricUseCasesParams {
	return factory.MetricUseCasesParams{
		MetricRepo:     p.metricRepo,
		MetricSvc:      p.metricSvc,
		Transactor:     p.transactor,
		MetricFileRepo: p.metricFileRepo,
		SyncFileWrites: p.syncFileWrites,
		AuditFileRepo:  p.auditFileRepo,
		AuditGateway:   p.auditGateway,
		AuditPublisher: p.auditPublisher,
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

// WithTransactor sets the transactor dependency.
func WithTransactor(t port.Transactor) option.Option[factoryParams] {
	return func(p *factoryParams) { p.transactor = t }
}

// WithMetricFileRepo sets the file snapshot repository.
func WithMetricFileRepo(r port.MetricFileRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.metricFileRepo = r }
}

// WithSyncFileWrites sets synchronous snapshot writes from mutation use cases.
func WithSyncFileWrites(enabled bool) option.Option[factoryParams] {
	return func(p *factoryParams) { p.syncFileWrites = enabled }
}

// WithAuditFileRepo sets the audit file repository.
func WithAuditFileRepo(r port.AuditFileRepository) option.Option[factoryParams] {
	return func(p *factoryParams) { p.auditFileRepo = r }
}

// WithAuditGateway sets the audit gateway.
func WithAuditGateway(g port.AuditGateway) option.Option[factoryParams] {
	return func(p *factoryParams) { p.auditGateway = g }
}

// WithAuditPublisher sets the audit publisher.
func WithAuditPublisher(pub port.AuditPublisher) option.Option[factoryParams] {
	return func(p *factoryParams) { p.auditPublisher = pub }
}
