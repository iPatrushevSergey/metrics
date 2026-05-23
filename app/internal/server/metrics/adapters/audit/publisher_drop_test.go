package audit

import (
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	portmocks "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuditEventPublisher_publishQueueFull(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	p := NewAuditEventPublisher(log, 1)
	_, err := p.Subscribe("file")
	require.NoError(t, err)

	p.Publish(dto.AuditEvent{TS: 1, Metrics: []string{"a"}})
	p.Publish(dto.AuditEvent{TS: 2, Metrics: []string{"b"}})
}
