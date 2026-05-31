package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPollGopsutilWorker_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := mocks.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	var ticks atomic.Int32
	factory := stubUseCaseFactory{
		pollGopsutil: stubUseCase{onExecute: func(context.Context, struct{}) (int, error) {
			ticks.Add(1)
			return 1, nil
		}},
	}

	w := NewPollGopsutilWorker(factory, log, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Run(ctx)
	time.Sleep(5 * time.Millisecond)
	cancel()
	time.Sleep(2 * time.Millisecond)

	assert.GreaterOrEqual(t, ticks.Load(), int32(1))
}
