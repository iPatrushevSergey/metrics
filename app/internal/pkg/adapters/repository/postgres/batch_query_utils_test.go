package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSendBatchQuery_empty(t *testing.T) {
	n, err := BuildSendBatchQuery[int](context.Background(), nil, "", "", 1, nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestBuildSendBatchQuery_invalidParams(t *testing.T) {
	_, err := BuildSendBatchQuery[int](context.Background(), nil, "", "", 0, []int{1}, nil, nil)
	assert.Error(t, err)
}
