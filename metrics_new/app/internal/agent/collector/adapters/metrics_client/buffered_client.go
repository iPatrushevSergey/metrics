package metrics_client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
)

// job is one pooled outbound call: requestMethod runs the concrete MetricsClient method.
type job struct {
	requestMethod func(context.Context) error
}

// BufferedClient is a MetricsClient decorator: a worker pool with two channels: jobs and results.
type BufferedClient struct {
	metricsClient port.MetricsClient
	log           port.Logger
	workers       int
	sendCtx       context.Context
	jobs          chan job
	results       chan error
	mu            sync.Mutex
	started       bool
	poolWg        sync.WaitGroup
	resultsWg     sync.WaitGroup
}

var _ port.MetricsClient = (*BufferedClient)(nil)

// NewBufferedClient initializes a new BufferedClient.
func NewBufferedClient(
	metricsClient port.MetricsClient,
	log port.Logger,
	workers int,
	sendCtx context.Context,
) (*BufferedClient, error) {
	if workers <= 0 {
		return nil, errors.New("metrics_client: workers must be > 0")
	}
	return &BufferedClient{
		metricsClient: metricsClient,
		log:           log,
		workers:       workers,
		sendCtx:       sendCtx,
	}, nil
}

// Start launches the results processor, then worker goroutines.
func (c *BufferedClient) Start() {
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
func (c *BufferedClient) processResults() {
	defer c.resultsWg.Done()
	for err := range c.results {
		if err != nil && !errors.Is(err, context.Canceled) {
			c.log.Error("metrics client batch send failed", "error", err)
		}
	}
}

// worker reads jobs from the jobs channel and runs the requestMethod.
func (c *BufferedClient) worker() {
	defer c.poolWg.Done()
	for j := range c.jobs {
		err := j.requestMethod(c.sendCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			c.results <- err
		}
	}
}

// Stop closes the jobs and results channels, waits for the workers to finish.
func (c *BufferedClient) Stop() {
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
func (c *BufferedClient) MetricsUpdateBatch(ctx context.Context, metrics []dto.MetricUpdateInput) error {
	j := job{
		requestMethod: func(callCtx context.Context) error {
			return c.metricsClient.MetricsUpdateBatch(callCtx, metrics)
		},
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.jobs <- j:
		return nil
	case <-time.After(1 * time.Second):
		return errors.New("metrics_client: jobs channel full, enqueue timed out")
	}
}
