package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/stretchr/testify/require"
)

type recordingSaver struct {
	saves int
	err   error
	last  map[string]model.Metric
}

func (s *recordingSaver) Save(metrics map[string]model.Metric) error {
	s.saves++
	s.last = metrics
	return s.err
}

// fakeRepo implements MetricRepository for error-path checks.
type fakeRepo struct {
	db             map[string]model.Metric
	createErr      error
	getAllErr      error
	createBatchErr error
	updateErr      error
	updateBatchErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{db: make(map[string]model.Metric)}
}

func (f *fakeRepo) GetByID(ctx context.Context, id string) (model.Metric, error) {
	m, ok := f.db[id]
	if !ok {
		return model.Metric{}, ErrNotFound
	}
	return m, nil
}

func (f *fakeRepo) GetByIDs(ctx context.Context, ids []string) (map[string]model.Metric, error) {
	out := make(map[string]model.Metric)
	for _, id := range ids {
		if m, ok := f.db[id]; ok {
			out[id] = m
		}
	}
	return out, nil
}

func (f *fakeRepo) GetAll(ctx context.Context) (map[string]model.Metric, error) {
	if f.getAllErr != nil {
		return nil, f.getAllErr
	}
	cp := make(map[string]model.Metric, len(f.db))
	for k, v := range f.db {
		cp[k] = v
	}
	return cp, nil
}

func (f *fakeRepo) Create(ctx context.Context, metric model.Metric) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.db[metric.ID] = metric
	return nil
}

func (f *fakeRepo) CreateBatch(ctx context.Context, metrics []model.Metric) error {
	if f.createBatchErr != nil {
		return f.createBatchErr
	}
	for _, m := range metrics {
		f.db[m.ID] = m
	}
	return nil
}

func (f *fakeRepo) Update(ctx context.Context, id string, metric model.Metric) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if _, ok := f.db[id]; !ok {
		return ErrNotFound
	}
	f.db[id] = metric
	return nil
}

func (f *fakeRepo) UpdateBatch(ctx context.Context, metrics []model.Metric) error {
	if f.updateBatchErr != nil {
		return f.updateBatchErr
	}
	for _, m := range metrics {
		if _, ok := f.db[m.ID]; !ok {
			return ErrNotFound
		}
		f.db[m.ID] = m
	}
	return nil
}

func (f *fakeRepo) Ping(ctx context.Context) error {
	return nil
}

func TestNewSyncFileRepository_Create_thenSave(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	saver := &recordingSaver{}
	r := NewSyncFileRepository(repo, saver)

	m := model.Metric{ID: "a", MType: model.Counter}
	require.NoError(t, r.Create(ctx, m))
	require.Equal(t, 1, saver.saves)
	require.Contains(t, saver.last, "a")
}

func TestNewSyncFileRepository_Update_thenSave(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	saver := &recordingSaver{}
	r := NewSyncFileRepository(repo, saver)

	require.NoError(t, r.Create(ctx, model.Metric{ID: "x", MType: model.Gauge}))
	saver.saves = 0

	require.NoError(t, r.Update(ctx, "x", model.Metric{ID: "x", MType: model.Counter}))
	require.Equal(t, 1, saver.saves)
	require.Equal(t, model.Counter, saver.last["x"].MType)
}

func TestNewSyncFileRepository_CreateBatch_thenSave(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	saver := &recordingSaver{}
	r := NewSyncFileRepository(repo, saver)

	require.NoError(t, r.CreateBatch(ctx, []model.Metric{
		{ID: "1", MType: model.Gauge},
		{ID: "2", MType: model.Counter},
	}))
	require.Equal(t, 1, saver.saves)
	require.Len(t, saver.last, 2)
}

func TestNewSyncFileRepository_UpdateBatch_thenSave(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	saver := &recordingSaver{}
	r := NewSyncFileRepository(repo, saver)

	require.NoError(t, r.Create(ctx, model.Metric{ID: "u", MType: model.Gauge}))
	saver.saves = 0

	require.NoError(t, r.UpdateBatch(ctx, []model.Metric{{ID: "u", MType: model.Counter}}))
	require.Equal(t, 1, saver.saves)
	require.Equal(t, model.Counter, saver.last["u"].MType)
}

func TestNewSyncFileRepository_GetByID_delegates(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	saver := &recordingSaver{}
	r := NewSyncFileRepository(repo, saver)

	require.NoError(t, r.Create(ctx, model.Metric{ID: "q", MType: model.Gauge}))
	saver.saves = 0

	got, err := r.GetByID(ctx, "q")
	require.NoError(t, err)
	require.Equal(t, "q", got.ID)
	require.Zero(t, saver.saves, "read-only ops must not sync to file")
}

func TestNewSyncFileRepository_GetByIDs_delegates(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	saver := &recordingSaver{}
	r := NewSyncFileRepository(repo, saver)

	require.NoError(t, r.Create(ctx, model.Metric{ID: "a", MType: model.Gauge}))
	saver.saves = 0

	out, err := r.GetByIDs(ctx, []string{"a", "missing"})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Zero(t, saver.saves)
}

func TestNewSyncFileRepository_GetAll_delegates(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	saver := &recordingSaver{}
	r := NewSyncFileRepository(repo, saver)

	require.NoError(t, r.Create(ctx, model.Metric{ID: "g", MType: model.Counter}))
	saver.saves = 0

	all, err := r.GetAll(ctx)
	require.NoError(t, err)
	require.Contains(t, all, "g")
	require.Zero(t, saver.saves)
}

func TestNewSyncFileRepository_Ping_delegates(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	r := NewSyncFileRepository(repo, &recordingSaver{})
	require.NoError(t, r.Ping(ctx))
}

func TestNewSyncFileRepository_Create_repoError_noSave(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.createErr = errors.New("create failed")
	saver := &recordingSaver{}
	r := NewSyncFileRepository(repo, saver)

	err := r.Create(ctx, model.Metric{ID: "x"})
	require.Error(t, err)
	require.Zero(t, saver.saves)
}

func TestNewSyncFileRepository_Create_getAllError(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.getAllErr = errors.New("getall failed")
	saver := &recordingSaver{}
	r := NewSyncFileRepository(repo, saver)

	err := r.Create(ctx, model.Metric{ID: "only"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "getall")
	require.Zero(t, saver.saves)
}

func TestNewSyncFileRepository_Create_saverError(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	saver := &recordingSaver{err: errors.New("disk full")}
	r := NewSyncFileRepository(repo, saver)

	err := r.Create(ctx, model.Metric{ID: "z"})
	require.Error(t, err)
}
