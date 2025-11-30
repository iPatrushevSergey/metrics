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

func (r *SyncFileRepository) Create(m model.Metric) error {
	if err := r.repo.Create(m); err != nil {
		return err
	}

	return r.fs.Save(r.repo.GetAll())
}

func (r *SyncFileRepository) Update(id string, m model.Metric) error {
	if err := r.repo.Update(id, m); err != nil {
		return err
	}

	return r.fs.Save(r.repo.GetAll())
}

func (r *SyncFileRepository) GetByID(id string) (model.Metric, bool) {
	return r.repo.GetByID(id)
}

func (r *SyncFileRepository) GetAll() map[string]model.Metric {
	return r.repo.GetAll()
}

func (r *SyncFileRepository) Ping(ctx context.Context) error {
	return r.repo.Ping(ctx)
}
