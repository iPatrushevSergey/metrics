package metricsgateway

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
)

// job is one pooled outbound call: requestMethod runs the concrete MetricsGateway method.
type job struct {
	requestMethod func(context.Context) error
}

// BufferedMetricsGateway wraps MetricsGateway with a fixed-size worker pool and bounded job queue.
type BufferedMetricsGateway struct {
	delegate  port.MetricsGateway
	log       port.Logger
	workers   int
	sendCtx   context.Context
	jobs      chan job
	results   chan error
	mu        sync.Mutex
	started   bool
	poolWg    sync.WaitGroup
	resultsWg sync.WaitGroup
}

var _ port.MetricsGateway = (*BufferedMetricsGateway)(nil)

// NewBufferedMetricsGateway builds a rate-limited MetricsGateway decorator.
func NewBufferedMetricsGateway(
	delegate port.MetricsGateway,
	log port.Logger,
	workers int,
	sendCtx context.Context,
) (*BufferedMetricsGateway, error) {
	if workers <= 0 {
		return nil, errors.New("metricsgateway: workers must be > 0")
	}
	return &BufferedMetricsGateway{
		delegate: delegate,
		log:      log,
		workers:  workers,
		sendCtx:  sendCtx,
	}, nil
}

// Start launches the results processor, then worker goroutines.
func (c *BufferedMetricsGateway) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		return
	}
	bufferSize := c.workers
	c.jobs = make(chan job, bufferSize)
	c.results = make(chan error, bufferSize)

	c.resultsWg.Add(1)
	go c.processResults()

	for w := 0; w < c.workers; w++ {
		c.poolWg.Add(1)
		go c.worker()
	}
	c.started = true
}

// processResults reads results from the results channel and logs errors.
func (c *BufferedMetricsGateway) processResults() {
	defer c.resultsWg.Done()
	for err := range c.results {
		if err != nil && !errors.Is(err, context.Canceled) {
			c.log.Error("metrics gateway batch send failed", "error", err)
		}
	}
}

// worker reads jobs from the jobs channel and runs the requestMethod.
func (c *BufferedMetricsGateway) worker() {
	defer c.poolWg.Done()
	for j := range c.jobs {
		err := j.requestMethod(c.sendCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			c.results <- err
		}
	}
}

// Stop closes the jobs and results channels, waits for the workers to finish.
func (c *BufferedMetricsGateway) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.started {
		return
	}
	close(c.jobs)
	c.poolWg.Wait()
	close(c.results)
	c.resultsWg.Wait()
	c.started = false
}

// MetricsUpdateBatch enqueues a batch send.
func (c *BufferedMetricsGateway) MetricsUpdateBatch(ctx context.Context, metrics []dto.MetricUpdateInput) error {
	j := job{
		requestMethod: func(callCtx context.Context) error {
			return c.delegate.MetricsUpdateBatch(callCtx, metrics)
		},
	}
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.jobs <- j:
		return nil
	case <-timer.C:
		c.log.Warn(
			"batch jobs channel is full, dropping batch to prevent goroutine backlog",
			"rate_limit", c.workers,
			"batch_size", len(metrics),
		)
		return fmt.Errorf("batch jobs channel is full, rate limit exceeded")
	}
}
