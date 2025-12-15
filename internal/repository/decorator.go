package repository

import (
	"context"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

// MetricsSaver interface for saving metrics to external storage
type MetricsSaver interface {
	Save(metrics map[string]model.Metric) error
}

// SyncFileRepository a decorator for synchronously saving metrics to a file with each change
type SyncFileRepository struct {
	repo  MetricRepository
	saver MetricsSaver
}

// NewSyncFileRepository creates a new instance of SyncFileRepository
func NewSyncFileRepository(
	repo MetricRepository,
	saver MetricsSaver,
) *SyncFileRepository {
	return &SyncFileRepository{
		repo:  repo,
		saver: saver,
	}
}

func (r *SyncFileRepository) Create(ctx context.Context, m model.Metric) error {
	if err := r.repo.Create(ctx, m); err != nil {
		return err
	}

	allMetrics, err := r.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	return r.saver.Save(allMetrics)
}

func (r *SyncFileRepository) Update(ctx context.Context, id string, m model.Metric) error {
	if err := r.repo.Update(ctx, id, m); err != nil {
		return err
	}

	allMetrics, err := r.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	return r.saver.Save(allMetrics)
}

func (r *SyncFileRepository) GetByID(ctx context.Context, id string) (model.Metric, error) {
	return r.repo.GetByID(ctx, id)
}

func (r *SyncFileRepository) GetByIDs(ctx context.Context, ids []string) (map[string]model.Metric, error) {
	return r.repo.GetByIDs(ctx, ids)
}

func (r *SyncFileRepository) GetAll(ctx context.Context) (map[string]model.Metric, error) {
	return r.repo.GetAll(ctx)
}

func (r *SyncFileRepository) Ping(ctx context.Context) error {
	return r.repo.Ping(ctx)
}

// CreateBatch creates several metrics and synchronously saves them to a file
func (r *SyncFileRepository) CreateBatch(ctx context.Context, metrics []model.Metric) error {
	if err := r.repo.CreateBatch(ctx, metrics); err != nil {
		return err
	}

	allMetrics, err := r.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	return r.saver.Save(allMetrics)
}

// UpdateBatch updates several metrics and synchronously saves them to a file
func (r *SyncFileRepository) UpdateBatch(ctx context.Context, metrics []model.Metric) error {
	if err := r.repo.UpdateBatch(ctx, metrics); err != nil {
		return err
	}

	allMetrics, err := r.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	return r.saver.Save(allMetrics)
}
