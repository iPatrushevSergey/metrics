package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSnapshotWorker_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := mocks.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	var ticks atomic.Int32
	factory := stubUseCaseFactory{
		metricsSnapshot: stubUseCase[struct{}, int]{
			onExecute: func(context.Context, struct{}) (int, error) {
				ticks.Add(1)
				return 1, nil
			},
		},
	}

	w := NewSnapshotWorker(factory, log, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Run(ctx)
	time.Sleep(15 * time.Millisecond)
	cancel()

	assert.GreaterOrEqual(t, ticks.Load(), int32(1))
}
