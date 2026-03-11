package audit

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/iPatrushevSergey/metrics/internal/logger"
	"go.uber.org/zap"
)

// Queue size is estimated as:
//
//	buffer ~= peak_rps * tolerated_sink_pause_seconds
//
// For the current default target of 500 RPS and up to 1 second
// of temporary sink slowdown, keep 500 events buffered per sink.
const defaultPublisherQueueSize = 500

type observerWorker struct {
	observer Observer
	name     string
	events   chan Event
}

// publisher sends events to a list of observers.
type publisher struct {
	workers []observerWorker
	log     logger.Logger
	wg      sync.WaitGroup
	mu      sync.RWMutex
	closed  bool

	workerCtx    context.Context
	workerCancel context.CancelFunc
}

// NewPublisher returns a Publisher. If no observers given, Notify is a no-op.
func NewPublisher(log logger.Logger, observers ...Observer) Publisher {
	workerCtx, workerCancel := context.WithCancel(context.Background())
	p := &publisher{
		log:          log,
		workerCtx:    workerCtx,
		workerCancel: workerCancel,
	}

	for _, observer := range observers {
		if observer == nil {
			continue
		}

		worker := observerWorker{
			observer: observer,
			name:     fmt.Sprintf("%T", observer),
			events:   make(chan Event, defaultPublisherQueueSize),
		}

		p.workers = append(p.workers, worker)
		p.wg.Add(1)
		go p.runWorker(worker)
	}

	return p
}

// Notify publishes audit event asynchronously.
func (p *publisher) Notify(e Event) {
	p.mu.RLock()
	if p.closed || len(p.workers) == 0 {
		p.mu.RUnlock()
		return
	}

	for _, worker := range p.workers {
		select {
		case worker.events <- cloneEvent(e):
		default:
			if p.log != nil {
				p.log.Warn(
					"audit event dropped",
					zap.String("observer", worker.name),
					zap.String("reason", "queue is full"),
				)
			}
		}
	}

	p.mu.RUnlock()
}

// Close waits for in-flight notifications.
func (p *publisher) Close(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}

	p.closed = true
	for _, worker := range p.workers {
		close(worker.events)
	}
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		p.workerCancel()
		return ctx.Err()
	}
}

func (p *publisher) runWorker(worker observerWorker) {
	defer p.wg.Done()

	for event := range worker.events {
		if err := worker.observer.Publish(p.workerCtx, event); err != nil && p.log != nil {
			p.log.Error(
				"audit observer error",
				zap.String("observer", worker.name),
				zap.Error(err),
			)
		}
	}
}

func cloneEvent(e Event) Event {
	return Event{
		TS:        e.TS,
		IPAddress: e.IPAddress,
		Metrics:   slices.Clone(e.Metrics),
	}
}
