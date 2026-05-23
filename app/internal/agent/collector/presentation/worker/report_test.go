package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	portmocks "github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestReportWorker_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	var ticks atomic.Int32
	factory := stubUseCaseFactory{
		reportBatch: stubUseCase{onExecute: func(context.Context, struct{}) (int, error) {
			ticks.Add(1)
			return 1, nil
		}},
	}

	sendCtx := context.Background()
	w := NewReportWorker(sendCtx, factory, log, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Run(ctx)
	time.Sleep(15 * time.Millisecond)
	cancel()
	w.Wait()

	assert.GreaterOrEqual(t, ticks.Load(), int32(1))
}
