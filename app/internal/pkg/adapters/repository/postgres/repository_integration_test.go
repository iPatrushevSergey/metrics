//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgreskit "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/retry"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/testutil"
	metricrepo "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/postgres"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
)

const (
	benchN200   = 200
	benchN2000  = 2000
	benchN12000 = 12000
)

func setupRepo(tb testing.TB) (*postgreskit.Transactor, *metricrepo.MetricPostgresRepository, *pgxpool.Pool) {
	tb.Helper()
	pool := testutil.SetupPostgres(tb)
	tx := postgreskit.NewTransactor(pool, retry.WithMaxRetries(0))
	return tx, metricrepo.NewMetricPostgresRepository(tx), pool
}

func truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `TRUNCATE metrics`)
	require.NoError(t, err)
}

func gauge(id string, v float64) entity.Metric {
	return entity.Metric{ID: id, MType: entity.Gauge, Value: &v, Hash: "hash"}
}

func gaugeBatch(n int, prefix string) []entity.Metric {
	out := make([]entity.Metric, n)
	for i := range out {
		v := float64(i)
		out[i] = gauge(fmt.Sprintf("%s%d", prefix, i), v)
	}
	return out
}

func TestMetricRepository_Ping(t *testing.T) {
	_, repo, _ := setupRepo(t)
	require.NoError(t, repo.Ping(context.Background()))
}

func TestMetricRepository_CreateAndGetByID(t *testing.T) {
	_, repo, _ := setupRepo(t)
	ctx := context.Background()
	m := gauge("m1", 1.5)

	require.NoError(t, repo.Create(ctx, m))

	found, err := repo.GetByID(ctx, "m1")
	require.NoError(t, err)
	assert.Equal(t, "m1", found.ID)
	assert.Equal(t, entity.Gauge, found.MType)
	require.NotNil(t, found.Value)
	assert.InDelta(t, 1.5, *found.Value, 0.01)
}

func TestMetricRepository_GetByID_NotFound(t *testing.T) {
	_, repo, _ := setupRepo(t)
	_, err := repo.GetByID(context.Background(), "missing")
	assert.ErrorIs(t, err, application.ErrNotFound)
}

func TestMetricRepository_GetByIDs(t *testing.T) {
	_, repo, _ := setupRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, gauge("a", 1)))
	require.NoError(t, repo.Create(ctx, gauge("b", 2)))

	got, err := repo.GetByIDs(ctx, []string{"a", "b", "c"})
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestMetricRepository_GetAll(t *testing.T) {
	_, repo, _ := setupRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, gauge("x", 10)))
	require.NoError(t, repo.Create(ctx, gauge("y", 20)))

	all, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestMetricRepository_Update(t *testing.T) {
	_, repo, _ := setupRepo(t)
	ctx := context.Background()
	m := gauge("u1", 1)

	require.NoError(t, repo.Create(ctx, m))
	v := 99.0
	m.Value = &v
	require.NoError(t, repo.Update(ctx, m))

	found, err := repo.GetByID(ctx, "u1")
	require.NoError(t, err)
	assert.InDelta(t, 99, *found.Value, 0.01)
}

func TestMetricRepository_Update_NotFound(t *testing.T) {
	_, repo, _ := setupRepo(t)
	err := repo.Update(context.Background(), gauge("ghost", 1))
	assert.ErrorIs(t, err, application.ErrNotFound)
}

func TestMetricRepository_CreateBatchWithParams(t *testing.T) {
	_, repo, pool := setupRepo(t)
	ctx := context.Background()
	truncate(t, pool)

	m := gaugeBatch(3, "params")
	require.NoError(t, repo.CreateBatchWithParams(ctx, m))

	var n int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM metrics WHERE id LIKE 'params%'`).Scan(&n))
	assert.Equal(t, 3, n)
}

func TestMetricRepository_CreateBatchWithUnnest(t *testing.T) {
	_, repo, pool := setupRepo(t)
	ctx := context.Background()
	truncate(t, pool)

	m := gaugeBatch(3, "unnest")
	require.NoError(t, repo.CreateBatchWithUnnest(ctx, m))

	var n int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM metrics WHERE id LIKE 'unnest%'`).Scan(&n))
	assert.Equal(t, 3, n)
}

func TestMetricRepository_CreateBatchWithCopy(t *testing.T) {
	_, repo, pool := setupRepo(t)
	ctx := context.Background()
	truncate(t, pool)

	m := gaugeBatch(3, "copy")
	require.NoError(t, repo.CreateBatchWithCopy(ctx, m))

	var n int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM metrics WHERE id LIKE 'copy%'`).Scan(&n))
	assert.Equal(t, 3, n)
}

func TestMetricRepository_CreateBatchWithPrepare(t *testing.T) {
	tx, repo, pool := setupRepo(t)
	ctx := context.Background()
	truncate(t, pool)

	m := gaugeBatch(3, "prep")
	err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
		return repo.CreateBatchWithPrepare(txCtx, m)
	})
	require.NoError(t, err)

	var n int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM metrics WHERE id LIKE 'prep%'`).Scan(&n))
	assert.Equal(t, 3, n)
}

func TestMetricRepository_UpdateBatchWithParams(t *testing.T) {
	_, repo, pool := setupRepo(t)
	ctx := context.Background()
	truncate(t, pool)

	m := gaugeBatch(3, "uparams")
	require.NoError(t, repo.CreateBatchWithUnnest(ctx, m))
	for i := range m {
		m[i].Hash = fmt.Sprintf("h%d", i)
	}
	require.NoError(t, repo.UpdateBatchWithParams(ctx, m))

	var hash string
	require.NoError(t, pool.QueryRow(ctx, `SELECT hash FROM metrics WHERE id = 'uparams0'`).Scan(&hash))
	assert.Equal(t, "h0", hash)
}

func TestMetricRepository_UpdateBatchWithUnnest(t *testing.T) {
	_, repo, pool := setupRepo(t)
	ctx := context.Background()
	truncate(t, pool)

	m := gaugeBatch(3, "uunnest")
	require.NoError(t, repo.CreateBatchWithUnnest(ctx, m))
	for i := range m {
		m[i].Hash = fmt.Sprintf("h%d", i)
	}
	require.NoError(t, repo.UpdateBatchWithUnnest(ctx, m))

	var hash string
	require.NoError(t, pool.QueryRow(ctx, `SELECT hash FROM metrics WHERE id = 'uunnest0'`).Scan(&hash))
	assert.Equal(t, "h0", hash)
}

func TestMetricRepository_UpdateBatchWithCopy(t *testing.T) {
	tx, repo, pool := setupRepo(t)
	ctx := context.Background()
	truncate(t, pool)

	m := gaugeBatch(3, "ucopy")
	require.NoError(t, repo.CreateBatchWithUnnest(ctx, m))
	for i := range m {
		m[i].Hash = fmt.Sprintf("h%d", i)
	}
	err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
		return repo.UpdateBatchWithCopy(txCtx, m)
	})
	require.NoError(t, err)

	var hash string
	require.NoError(t, pool.QueryRow(ctx, `SELECT hash FROM metrics WHERE id = 'ucopy0'`).Scan(&hash))
	assert.Equal(t, "h0", hash)
}

func TestMetricRepository_UpdateBatchWithPrepare(t *testing.T) {
	tx, repo, pool := setupRepo(t)
	ctx := context.Background()
	truncate(t, pool)

	m := gaugeBatch(3, "uprep")
	require.NoError(t, repo.CreateBatchWithUnnest(ctx, m))
	for i := range m {
		m[i].Hash = fmt.Sprintf("h%d", i)
	}
	err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
		return repo.UpdateBatchWithPrepare(txCtx, m)
	})
	require.NoError(t, err)

	var hash string
	require.NoError(t, pool.QueryRow(ctx, `SELECT hash FROM metrics WHERE id = 'uprep0'`).Scan(&hash))
	assert.Equal(t, "h0", hash)
}

// --- benchmarks ---

func truncateBench(tb testing.TB, pool *pgxpool.Pool) {
	tb.Helper()
	if _, err := pool.Exec(context.Background(), `TRUNCATE metrics`); err != nil {
		tb.Fatal(err)
	}
}

func BenchmarkCreateBatchParams_200(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN200, "bparams")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		if err := repo.CreateBatchWithParams(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchUnnest_200(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN200, "bunnest")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchCopy_200(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN200, "bcopy")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		if err := repo.CreateBatchWithCopy(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchPrepare_200(b *testing.B) {
	tx, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN200, "bprep")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
			return repo.CreateBatchWithPrepare(txCtx, metrics)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchParams_2000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN2000, "bparams")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		if err := repo.CreateBatchWithParams(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchUnnest_2000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN2000, "bunnest")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchCopy_2000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN2000, "bcopy")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		if err := repo.CreateBatchWithCopy(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchPrepare_2000(b *testing.B) {
	tx, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN2000, "bprep")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
			return repo.CreateBatchWithPrepare(txCtx, metrics)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchParams_12000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN12000, "bparams")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		if err := repo.CreateBatchWithParams(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchUnnest_12000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN12000, "bunnest")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchCopy_12000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN12000, "bcopy")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		if err := repo.CreateBatchWithCopy(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateBatchPrepare_12000(b *testing.B) {
	tx, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN12000, "bprep")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateBench(b, pool)
		err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
			return repo.CreateBatchWithPrepare(txCtx, metrics)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchParams_200(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN200, "buparams")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		if err := repo.UpdateBatchWithParams(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchUnnest_200(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN200, "buunnest")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		if err := repo.UpdateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchCopy_200(b *testing.B) {
	tx, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN200, "bucopy")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
			return repo.UpdateBatchWithCopy(txCtx, metrics)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchPrepare_200(b *testing.B) {
	tx, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN200, "buprep")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
			return repo.UpdateBatchWithPrepare(txCtx, metrics)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchParams_2000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN2000, "buparams")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		if err := repo.UpdateBatchWithParams(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchUnnest_2000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN2000, "buunnest")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		if err := repo.UpdateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchCopy_2000(b *testing.B) {
	tx, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN2000, "bucopy")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
			return repo.UpdateBatchWithCopy(txCtx, metrics)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchPrepare_2000(b *testing.B) {
	tx, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN2000, "buprep")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
			return repo.UpdateBatchWithPrepare(txCtx, metrics)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchParams_12000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN12000, "buparams")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		if err := repo.UpdateBatchWithParams(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchUnnest_12000(b *testing.B) {
	_, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN12000, "buunnest")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		if err := repo.UpdateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchCopy_12000(b *testing.B) {
	tx, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN12000, "bucopy")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
			return repo.UpdateBatchWithCopy(txCtx, metrics)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchPrepare_12000(b *testing.B) {
	tx, repo, pool := setupRepo(b)
	ctx := context.Background()
	metrics := gaugeBatch(benchN12000, "buprep")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		truncateBench(b, pool)
		if err := repo.CreateBatchWithUnnest(ctx, metrics); err != nil {
			b.Fatal(err)
		}
		for j := range metrics {
			metrics[j].Hash = fmt.Sprintf("h%d", j)
		}
		b.StartTimer()

		err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
			return repo.UpdateBatchWithPrepare(txCtx, metrics)
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
