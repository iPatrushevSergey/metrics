package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// maxQueryParams is the maximum number of parameters that can be passed to a query.
const maxQueryParams = 65535

// TakeRowParams takes bind values for one row into args.
type TakeRowParams[T any] func(item T, args []any)

// BuildSendBatchQuery builds a query to send a batch of items to the database.
func BuildSendBatchQuery[T any](
	ctx context.Context,
	q Querier,
	SQLQueryHead string,
	SQLQueryTail string,
	paramsInRow int,
	items []T,
	takeRowParams TakeRowParams[T],
	paramCasts []string,
) (int64, error) {
	if len(items) == 0 {
		return 0, nil
	}
	if paramsInRow <= 0 {
		return 0, fmt.Errorf("BuildSendBatchQuery: paramsInRow must be positive")
	}
	if len(paramCasts) != 0 && len(paramCasts) != paramsInRow {
		return 0, fmt.Errorf("BuildSendBatchQuery: paramCasts length %d, want 0 or %d", len(paramCasts), paramsInRow)
	}

	chunkSize := maxQueryParams / paramsInRow
	maxRowsInChunk := chunkSize
	if n := len(items); n < maxRowsInChunk {
		maxRowsInChunk = n
	}

	var b strings.Builder
	b.Grow(len(SQLQueryHead) + len(SQLQueryTail) + maxRowsInChunk*paramsInRow*20)
	var tmpParamNum [5]byte

	chunkParams := make([]any, maxRowsInChunk*paramsInRow)

	var totalAffected int64
	for start := 0; start < len(items); start += chunkSize {
		end := start + chunkSize
		if end > len(items) {
			end = len(items)
		}
		chunkLen := end - start
		currParams := chunkParams[:chunkLen*paramsInRow]

		b.Reset()
		b.WriteString(SQLQueryHead)

		for i := start; i < end; i++ {
			if i > start {
				b.WriteByte(',')
			}
			lenOffset := (i - start) * paramsInRow
			takeRowParams(items[i], currParams[lenOffset:lenOffset+paramsInRow])

			startParamNumber := lenOffset + 1
			b.WriteByte('(')
			for paramNumber := 0; paramNumber < paramsInRow; paramNumber++ {
				if paramNumber > 0 {
					b.WriteByte(',')
				}
				b.WriteByte('$')
				b.Write(strconv.AppendInt(tmpParamNum[:0], int64(startParamNumber+paramNumber), 10))
				if len(paramCasts) > 0 {
					b.WriteString("::")
					b.WriteString(paramCasts[paramNumber])
				}
			}
			b.WriteByte(')')
		}

		b.WriteString(SQLQueryTail)

		commandTag, err := q.Exec(ctx, b.String(), currParams...)
		if err != nil {
			return totalAffected, err
		}
		totalAffected += commandTag.RowsAffected()
	}
	return totalAffected, nil
}

// TakeRow returns values for one COPY row.
type TakeRow[T any] func(item T) []any

// NewCopyFromSource streams items into pgx.CopyFrom.
func NewCopyFromSource[T any](items []T, take TakeRow[T]) pgx.CopyFromSource {
	return &copyFromSource[T]{items: items, take: take}
}

// copyFromSource streams items into pgx.CopyFrom.
type copyFromSource[T any] struct {
	items []T
	take  TakeRow[T]
	idx   int
	err   error
}

// Next returns true if there are more rows to copy.
func (s *copyFromSource[T]) Next() bool {
	return s.idx < len(s.items) && s.err == nil
}

// Values returns the next row to copy.
func (s *copyFromSource[T]) Values() ([]any, error) {
	if s.idx >= len(s.items) {
		return nil, fmt.Errorf("copyFromSource: no more rows")
	}
	row := s.take(s.items[s.idx])
	s.idx++
	return row, nil
}

// Err returns the error if any.
func (s *copyFromSource[T]) Err() error {
	return s.err
}
