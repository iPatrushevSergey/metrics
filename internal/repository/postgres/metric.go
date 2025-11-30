package postgres

import (
	"context"
	"database/sql"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
)

type PostgresMetricRepository struct {
	db *sql.DB
}

func NewPostgresMetricRepository(db *sql.DB) repository.MetricRepository {
	return &PostgresMetricRepository{
		db: db,
	}
}

func (r *PostgresMetricRepository) GetByID(id string) (model.Metric, bool) {
	return model.Metric{}, false
}

func (r *PostgresMetricRepository) GetAll() map[string]model.Metric {
	return map[string]model.Metric{}
}

func (r *PostgresMetricRepository) Create(metric model.Metric) error {
	return nil
}

func (r *PostgresMetricRepository) Update(id string, metric model.Metric) error {
	return nil
}

func (r *PostgresMetricRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
