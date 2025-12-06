package postgres

import (
	"context"
	"database/sql"

	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
	"go.uber.org/zap"
)

type PostgresMetricRepository struct {
	db *sql.DB
}

func NewPostgresMetricRepository(db *sql.DB) repository.MetricRepository {
	return &PostgresMetricRepository{
		db: db,
	}
}

func (r *PostgresMetricRepository) GetByID(ctx context.Context, id string) (model.Metric, bool) {
	query := `SELECT id, mtype, delta, value FROM metrics WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var m model.Metric

	err := row.Scan(&m.ID, &m.MType, &m.Delta, &m.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Metric{}, false
		}
		logger.Log.Error("error in GetByID", zap.Error(err))
		return model.Metric{}, false
	}

	return m, true
}

func (r *PostgresMetricRepository) GetAll(ctx context.Context) map[string]model.Metric {
	metrics := make(map[string]model.Metric)
	query := `SELECT id, mtype, delta, value FROM metrics`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Log.Error("error in GetAll", zap.Error(err))
		return metrics
	}
	defer rows.Close()

	for rows.Next() {
		var m model.Metric
		if err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value); err != nil {
			logger.Log.Error("error in getall during iteration", zap.Error(err))
			continue
		}
		metrics[m.ID] = m
	}

	if err := rows.Err(); err != nil {
		logger.Log.Error("error in getall during the final check", zap.Error(err))
		return metrics
	}

	return metrics
}

func (r *PostgresMetricRepository) Create(ctx context.Context, metric model.Metric) error {
	query := `INSERT INTO metrics (id, mtype, delta, value) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value)
	return err
}

func (r *PostgresMetricRepository) Update(ctx context.Context, id string, metric model.Metric) error {
	query := `UPDATE metrics SET delta = $1, value = $2 WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, metric.Delta, metric.Value, id)
	return err
}

func (r *PostgresMetricRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
