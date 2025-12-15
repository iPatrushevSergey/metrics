package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
)

// PostgresMetricRepository implements MetricRepository for PostgreSQL
type PostgresMetricRepository struct {
	db *sql.DB
}

// NewPostgresMetricRepository creates a new instance of PostgresMetricRepository
func NewPostgresMetricRepository(db *sql.DB) repository.MetricRepository {
	return &PostgresMetricRepository{
		db: db,
	}
}

func (r *PostgresMetricRepository) GetByID(ctx context.Context, id string) (model.Metric, error) {
	query := `SELECT id, mtype, delta, value, hash FROM metrics WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var m model.Metric

	err := row.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Metric{}, repository.ErrNotFound
		}
		return model.Metric{}, err
	}

	return m, nil
}

// GetByIDs returns metrics based on a list of IDs
func (r *PostgresMetricRepository) GetByIDs(ctx context.Context, ids []string) (map[string]model.Metric, error) {
	result := make(map[string]model.Metric, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	query := `SELECT id, mtype, delta, value, hash FROM metrics WHERE id = ANY($1)`
	rows, err := r.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m model.Metric
		if err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash); err != nil {
			return nil, err
		}
		result[m.ID] = m
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresMetricRepository) GetAll(ctx context.Context) (map[string]model.Metric, error) {
	metrics := make(map[string]model.Metric)
	query := `SELECT id, mtype, delta, value, hash FROM metrics`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m model.Metric
		if err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash); err != nil {
			return nil, err
		}
		metrics[m.ID] = m
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (r *PostgresMetricRepository) Create(ctx context.Context, metric model.Metric) error {
	_, err := r.GetByID(ctx, metric.ID)
	if err == nil {
		return repository.ErrAlreadyExists
	}
	if err != repository.ErrNotFound {
		return err
	}

	query := `INSERT INTO metrics (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5)`
	_, err = r.db.ExecContext(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash)
	if err != nil {
		return err
	}
	return nil
}

// CreateBatch creates multiple metrics in a single transaction
func (r *PostgresMetricRepository) CreateBatch(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO metrics (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, metric := range metrics {
		if _, err := stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *PostgresMetricRepository) Update(ctx context.Context, id string, metric model.Metric) error {
	query := `UPDATE metrics SET mtype = $1, delta = $2, value = $3, hash = $4 WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, metric.MType, metric.Delta, metric.Value, metric.Hash, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *PostgresMetricRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// UpdateBatch updates multiple metrics in a single transaction
func (r *PostgresMetricRepository) UpdateBatch(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `UPDATE metrics SET mtype = $1, delta = $2, value = $3, hash = $4 WHERE id = $5`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, metric := range metrics {
		result, err := stmt.ExecContext(ctx, metric.MType, metric.Delta, metric.Value, metric.Hash, metric.ID)
		if err != nil {
			tx.Rollback()
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}

		if rowsAffected == 0 {
			tx.Rollback()
			return repository.ErrNotFound
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
