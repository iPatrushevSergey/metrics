package audit

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/stretchr/testify/require"
)

func TestAuditFileRepository_append(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "audit.jsonl")

	repo, err := NewAuditFileRepository(path)
	require.NoError(t, err)

	event := dto.AuditEvent{TS: 1, Metrics: []string{"cpu"}, IPAddress: "127.0.0.1"}
	require.NoError(t, repo.Append(ctx, event))
	require.NoError(t, repo.Close())
}
