package usecase

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRecordAuditToFile_Execute(t *testing.T) {
	ctx := context.Background()
	event := dto.AuditEvent{TS: 1, Metrics: []string{"cpu"}, IPAddress: "127.0.0.1"}

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockAuditFileRepository(ctrl)
	repo.EXPECT().Append(ctx, event).Return(nil)

	uc := NewRecordAuditToFile(repo)
	_, err := uc.Execute(ctx, event)
	require.NoError(t, err)
}
