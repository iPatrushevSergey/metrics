package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuditFileSubscriber_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := mocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	var handled atomic.Int32
	factory := stubUseCaseFactory{
		recordAudit: stubUseCase[dto.AuditEvent, struct{}]{
			onExecute: func(context.Context, dto.AuditEvent) (struct{}, error) {
				handled.Add(1)
				return struct{}{}, nil
			},
		},
	}

	events := make(chan dto.AuditEvent, 1)
	events <- dto.AuditEvent{TS: 1}
	close(events)

	w := NewAuditFileSubscriber(events, factory, log)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	w.Run(ctx)

	assert.Equal(t, int32(1), handled.Load())
}
