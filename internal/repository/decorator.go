package repository

import (
	"context"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

// MetricsSaver интерфейс для сохранения метрик во внешнее хранилище
type MetricsSaver interface {
	Save(metrics map[string]model.Metric) error
}

// SyncFileRepository декоратор для синхронного сохранения метрик в файл при каждом изменении
type SyncFileRepository struct {
	repo  MetricRepository
	saver MetricsSaver
}

// NewSyncFileRepository создает новый экземпляр SyncFileRepository
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

func (r *SyncFileRepository) GetAll(ctx context.Context) (map[string]model.Metric, error) {
	return r.repo.GetAll(ctx)
}

func (r *SyncFileRepository) Ping(ctx context.Context) error {
	return r.repo.Ping(ctx)
}
