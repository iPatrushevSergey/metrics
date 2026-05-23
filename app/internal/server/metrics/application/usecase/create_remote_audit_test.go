package usecase

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	portmocks "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateRemoteAudit_Execute(t *testing.T) {
	ctx := context.Background()
	event := dto.AuditEvent{TS: 1, Metrics: []string{"cpu"}}

	ctrl := gomock.NewController(t)
	gateway := portmocks.NewMockAuditGateway(ctrl)
	gateway.EXPECT().CreateAudit(ctx, event).Return(nil)

	uc := NewCreateRemoteAudit(gateway)
	_, err := uc.Execute(ctx, event)
	require.NoError(t, err)
}
