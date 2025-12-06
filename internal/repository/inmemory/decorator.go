package inmemory

import (
	"context"

	"github.com/iPatrushevSergey/metrics/internal/filestorage"
	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
)

type SyncFileRepository struct {
	repo repository.MetricRepository
	fs   *filestorage.FileStorage
}

func NewSyncFileRepository(
	repo repository.MetricRepository,
	fs *filestorage.FileStorage,
) *SyncFileRepository {
	return &SyncFileRepository{
		repo: repo,
		fs:   fs,
	}
}

func (r *SyncFileRepository) Create(ctx context.Context, m model.Metric) error {
	if err := r.repo.Create(ctx, m); err != nil {
		return err
	}

	return r.fs.Save(r.repo.GetAll(ctx))
}

func (r *SyncFileRepository) Update(ctx context.Context, id string, m model.Metric) error {
	if err := r.repo.Update(ctx, id, m); err != nil {
		return err
	}

	return r.fs.Save(r.repo.GetAll(ctx))
}

func (r *SyncFileRepository) GetByID(ctx context.Context, id string) (model.Metric, bool) {
	return r.repo.GetByID(ctx, id)
}

func (r *SyncFileRepository) GetAll(ctx context.Context) map[string]model.Metric {
	return r.repo.GetAll(ctx)
}

func (r *SyncFileRepository) Ping(ctx context.Context) error {
	return r.repo.Ping(ctx)
}
