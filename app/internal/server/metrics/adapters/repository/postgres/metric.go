package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	postgreskit "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/postgres/converter"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/postgres/model"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
)

// MetricPostgresRepository is a PostgreSQL implementation of the metric repository.
type MetricPostgresRepository struct {
	transactor *postgreskit.Transactor
	conv       converter.MetricConverter
}

// NewMetricPostgresRepository creates a new MetricPostgresRepository.
func NewMetricPostgresRepository(transactor *postgreskit.Transactor) *MetricPostgresRepository {
	return &MetricPostgresRepository{
		transactor: transactor,
		conv:       &converter.MetricConverterImpl{},
	}
}

// Ping checks database connectivity.
func (r *MetricPostgresRepository) Ping(ctx context.Context) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		return q.QueryRow(ctx, `SELECT 1`).Scan(new(int))
	})
}

// GetByID returns a metric by id.
func (r *MetricPostgresRepository) GetByID(ctx context.Context, id string) (entity.Metric, error) {
	SQLQuery := `SELECT id, mtype, delta, value, hash FROM metrics WHERE id = $1`
	var m entity.Metric

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, SQLQuery, id)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRow, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[model.Metric])
		if err != nil {
			return err
		}

		m = r.conv.ToEntity(dbRow)
		return nil
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Metric{}, application.ErrNotFound
		}
		return entity.Metric{}, err
	}

	return m, nil
}

// GetByIDs returns metrics for the given ids.
func (r *MetricPostgresRepository) GetByIDs(ctx context.Context, ids []string) (map[string]entity.Metric, error) {
	if len(ids) == 0 {
		return map[string]entity.Metric{}, nil
	}

	SQLQuery := `SELECT id, mtype, delta, value, hash FROM metrics WHERE id = ANY($1::text[])`

	metrics := make(map[string]entity.Metric)

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, SQLQuery, ids)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.Metric])
		if err != nil {
			return err
		}

		clear(metrics)
		for _, dbRow := range dbRows {
			entity := r.conv.ToEntity(dbRow)
			metrics[entity.ID] = entity
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// GetAll returns all metrics.
func (r *MetricPostgresRepository) GetAll(ctx context.Context) (map[string]entity.Metric, error) {
	SQLQuery := `SELECT id, mtype, delta, value, hash FROM metrics ORDER BY id`
	metrics := make(map[string]entity.Metric)

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, SQLQuery)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.Metric])
		if err != nil {
			return err
		}

		clear(metrics)
		for _, dbRow := range dbRows {
			entity := r.conv.ToEntity(dbRow)
			metrics[entity.ID] = entity
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// Create inserts a new metric row.
func (r *MetricPostgresRepository) Create(ctx context.Context, metric entity.Metric) error {
	SQLQuery := `INSERT INTO metrics (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5)`

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		dbMetric := r.conv.ToModel(metric)

		_, err := q.Exec(ctx, SQLQuery, dbMetric.ID, dbMetric.MType, dbMetric.Delta, dbMetric.Value, dbMetric.Hash)
		return err
	})
}

// CreateBatchWithParams inserts multiple metrics via VALUES placeholders.
func (r *MetricPostgresRepository) CreateBatchWithParams(ctx context.Context, metrics []entity.Metric) error {
	SQLQueryHead := `INSERT INTO metrics (id, mtype, delta, value, hash) VALUES `
	paramsInRow := 5

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		_, err := postgreskit.BuildSendBatchQuery(
			ctx, q, SQLQueryHead, "", paramsInRow, metrics,
			func(m entity.Metric, params []any) {
				dbMetric := r.conv.ToModel(m)
				params[0] = dbMetric.ID
				params[1] = dbMetric.MType
				params[2] = dbMetric.Delta
				params[3] = dbMetric.Value
				params[4] = dbMetric.Hash
			},
			nil,
		)
		return err
	})
}

// CreateBatchWithPrepare inserts multiple metrics via a prepared statement and pgx batch.
func (r *MetricPostgresRepository) CreateBatchWithPrepare(ctx context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	stmtName := "metrics_batch_insert"
	SQLQuery := `INSERT INTO metrics (id, mtype, delta, value, hash) VALUES ($1, $2, $3, $4, $5)`

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		tx, ok := q.(pgx.Tx)
		if !ok {
			return fmt.Errorf("CreateBatchWithPrepare: must run inside transaction")
		}
		if _, err := tx.Prepare(ctx, stmtName, SQLQuery); err != nil {
			return err
		}

		batch := &pgx.Batch{}
		for _, m := range metrics {
			dbMetric := r.conv.ToModel(m)
			batch.Queue(stmtName, dbMetric.ID, dbMetric.MType, dbMetric.Delta, dbMetric.Value, dbMetric.Hash)
		}

		batchResult := tx.SendBatch(ctx, batch)
		defer batchResult.Close()

		for range metrics {
			if _, err := batchResult.Exec(); err != nil {
				return err
			}
		}
		return batchResult.Close()
	})
}

// CreateBatchWithUnnest inserts multiple metrics via UNNEST arrays.
func (r *MetricPostgresRepository) CreateBatchWithUnnest(ctx context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	SQLQuery := `INSERT INTO metrics (id, mtype, delta, value, hash)
	SELECT id, mtype, delta, value, hash FROM UNNEST(
		$1::text[],
		$2::text[],
		$3::bigint[],
		$4::double precision[],
		$5::text[]
	) AS t(id, mtype, delta, value, hash)`

	ids := make([]string, 0, len(metrics))
	mtypes := make([]string, 0, len(metrics))
	deltas := make([]*int64, 0, len(metrics))
	values := make([]*float64, 0, len(metrics))
	hashes := make([]*string, 0, len(metrics))

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		ids = ids[:0]
		mtypes = mtypes[:0]
		deltas = deltas[:0]
		values = values[:0]
		hashes = hashes[:0]

		for _, m := range metrics {
			dbMetric := r.conv.ToModel(m)
			ids = append(ids, dbMetric.ID)
			mtypes = append(mtypes, dbMetric.MType)
			deltas = append(deltas, dbMetric.Delta)
			values = append(values, dbMetric.Value)
			hashes = append(hashes, dbMetric.Hash)
		}

		_, err := q.Exec(ctx, SQLQuery, ids, mtypes, deltas, values, hashes)
		return err
	})
}

// CreateBatchWithCopy inserts multiple metrics via COPY.
func (r *MetricPostgresRepository) CreateBatchWithCopy(ctx context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	targetTable := pgx.Identifier{"metrics"}
	targetColumns := []string{"id", "mtype", "delta", "value", "hash"}

	takeRow := func(m entity.Metric) []any {
		dbMetric := r.conv.ToModel(m)
		return []any{dbMetric.ID, dbMetric.MType, dbMetric.Delta, dbMetric.Value, dbMetric.Hash}
	}

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		_, err := q.CopyFrom(ctx, targetTable, targetColumns, postgreskit.NewCopyFromSource(metrics, takeRow))
		return err
	})
}

// Update updates an existing metric row.
func (r *MetricPostgresRepository) Update(ctx context.Context, metric entity.Metric) error {
	SQLQuery := `UPDATE metrics SET mtype = $2, delta = $3, value = $4, hash = $5 WHERE id = $1`

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		dbMetric := r.conv.ToModel(metric)

		commandTag, err := q.Exec(ctx, SQLQuery, dbMetric.ID, dbMetric.MType, dbMetric.Delta, dbMetric.Value, dbMetric.Hash)
		if err != nil {
			return err
		}
		if commandTag.RowsAffected() == 0 {
			return application.ErrNotFound
		}
		return nil
	})
}

// UpdateBatchWithParams updates multiple metrics via VALUES placeholders.
func (r *MetricPostgresRepository) UpdateBatchWithParams(ctx context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	SQLQueryHead := `UPDATE metrics AS m 
	SET mtype = t.mtype, delta = t.delta, value = t.value, hash = t.hash 
	FROM (VALUES `

	SQLQueryTail := `) AS t(id, mtype, delta, value, hash) 
	WHERE m.id = t.id`

	SQLQueryParamTypes := []string{
		"varchar(50)", "varchar(10)", "bigint", "double precision", "varchar(64)",
	}

	paramsInRow := 5

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		affected, err := postgreskit.BuildSendBatchQuery(
			ctx, q, SQLQueryHead, SQLQueryTail, paramsInRow, metrics,
			func(m entity.Metric, params []any) {
				dbMetric := r.conv.ToModel(m)
				params[0] = dbMetric.ID
				params[1] = dbMetric.MType
				params[2] = dbMetric.Delta
				params[3] = dbMetric.Value
				params[4] = dbMetric.Hash
			},
			SQLQueryParamTypes,
		)
		if err != nil {
			return err
		}
		if affected != int64(len(metrics)) {
			return application.ErrNotFound
		}
		return nil
	})
}

// UpdateBatchWithPrepare updates multiple metrics via a prepared statement and pgx batch.
func (r *MetricPostgresRepository) UpdateBatchWithPrepare(ctx context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	stmtName := "metrics_batch_update"
	SQLQuery := `UPDATE metrics SET mtype = $2, delta = $3, value = $4, hash = $5 WHERE id = $1`

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		tx, ok := q.(pgx.Tx)
		if !ok {
			return fmt.Errorf("UpdateBatchWithPrepare: must run inside transaction")
		}
		if _, err := tx.Prepare(ctx, stmtName, SQLQuery); err != nil {
			return err
		}

		batch := &pgx.Batch{}
		for _, m := range metrics {
			dbMetric := r.conv.ToModel(m)
			batch.Queue(stmtName, dbMetric.ID, dbMetric.MType, dbMetric.Delta, dbMetric.Value, dbMetric.Hash)
		}

		batchResult := tx.SendBatch(ctx, batch)
		defer batchResult.Close()

		var affected int64
		for range metrics {
			tag, err := batchResult.Exec()
			if err != nil {
				return err
			}
			affected += tag.RowsAffected()
		}
		if err := batchResult.Close(); err != nil {
			return err
		}
		if affected != int64(len(metrics)) {
			return application.ErrNotFound
		}
		return nil
	})
}

// UpdateBatchWithUnnest updates multiple metrics via UNNEST arrays.
func (r *MetricPostgresRepository) UpdateBatchWithUnnest(ctx context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	SQLQuery := `UPDATE metrics AS m 
	SET mtype = t.mtype, delta = t.delta, value = t.value, hash = t.hash 
	FROM UNNEST(
		$1::text[],
		$2::text[],
		$3::bigint[],
		$4::double precision[],
		$5::text[]
	) AS t(id, mtype, delta, value, hash) 
	WHERE m.id = t.id`

	ids := make([]string, 0, len(metrics))
	mtypes := make([]string, 0, len(metrics))
	deltas := make([]*int64, 0, len(metrics))
	values := make([]*float64, 0, len(metrics))
	hashes := make([]*string, 0, len(metrics))

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		ids = ids[:0]
		mtypes = mtypes[:0]
		deltas = deltas[:0]
		values = values[:0]
		hashes = hashes[:0]

		for _, m := range metrics {
			dbMetric := r.conv.ToModel(m)
			ids = append(ids, dbMetric.ID)
			mtypes = append(mtypes, dbMetric.MType)
			deltas = append(deltas, dbMetric.Delta)
			values = append(values, dbMetric.Value)
			hashes = append(hashes, dbMetric.Hash)
		}

		commandTag, err := q.Exec(ctx, SQLQuery, ids, mtypes, deltas, values, hashes)
		if err != nil {
			return err
		}
		if commandTag.RowsAffected() != int64(len(metrics)) {
			return application.ErrNotFound
		}
		return nil
	})
}

// UpdateBatchWithCopy updates multiple metrics via COPY.
// Must be called only inside a transaction: temp staging is visible only in that DB session.
func (r *MetricPostgresRepository) UpdateBatchWithCopy(ctx context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	targetTable := pgx.Identifier{"metrics_temp"}
	targetColumns := []string{"id", "mtype", "delta", "value", "hash"}

	SQLQueryCreate := `CREATE TEMP TABLE metrics_temp (
		id VARCHAR(50) NOT NULL,
		mtype VARCHAR(10) NOT NULL,
		delta BIGINT,
		value DOUBLE PRECISION,
		hash VARCHAR(64)
	) ON COMMIT DROP`

	SQLQueryUpdate := `UPDATE metrics AS m 
	SET mtype = t.mtype, delta = t.delta, value = t.value, hash = t.hash 
	FROM metrics_temp AS t 
	WHERE m.id = t.id`

	takeRow := func(m entity.Metric) []any {
		dbMetric := r.conv.ToModel(m)
		return []any{dbMetric.ID, dbMetric.MType, dbMetric.Delta, dbMetric.Value, dbMetric.Hash}
	}

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		if _, err := q.Exec(ctx, SQLQueryCreate); err != nil {
			return err
		}
		if _, err := q.CopyFrom(ctx, targetTable, targetColumns, postgreskit.NewCopyFromSource(metrics, takeRow)); err != nil {
			return err
		}

		commandTag, err := q.Exec(ctx, SQLQueryUpdate)
		if err != nil {
			return err
		}
		if commandTag.RowsAffected() != int64(len(metrics)) {
			return application.ErrNotFound
		}
		return nil
	})
}
