package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubQuerier struct {
	tag     pgconn.CommandTag
	err     error
	lastSQL string
}

func (s *stubQuerier) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	s.lastSQL = sql
	return s.tag, s.err
}

func (s *stubQuerier) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("not used")
}

func (s *stubQuerier) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("not used")
}

func (s *stubQuerier) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	panic("not used")
}

func (s *stubQuerier) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	panic("not used")
}

func TestBuildSendBatchQuery_empty(t *testing.T) {
	n, err := BuildSendBatchQuery[int](context.Background(), nil, "", "", 1, nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestBuildSendBatchQuery_invalidParams(t *testing.T) {
	_, err := BuildSendBatchQuery[int](context.Background(), nil, "", "", 0, []int{1}, nil, nil)
	assert.Error(t, err)
}

func TestBuildSendBatchQuery_exec(t *testing.T) {
	q := &stubQuerier{tag: pgconn.NewCommandTag("INSERT 0 2")}
	n, err := BuildSendBatchQuery[int](
		context.Background(),
		q,
		"INSERT INTO t (a) VALUES ",
		" ON CONFLICT DO NOTHING",
		1,
		[]int{1, 2},
		func(item int, args []any) { args[0] = item },
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)
	assert.NotEmpty(t, q.lastSQL)
}

func TestNewCopyFromSource(t *testing.T) {
	src := NewCopyFromSource([]int{10, 20}, func(i int) []any { return []any{i} })
	require.True(t, src.Next())
	vals, err := src.Values()
	require.NoError(t, err)
	assert.Equal(t, []any{10}, vals)
	require.True(t, src.Next())
	vals, err = src.Values()
	require.NoError(t, err)
	assert.Equal(t, []any{20}, vals)
	assert.False(t, src.Next())
}
