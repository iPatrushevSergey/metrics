package postgres

import (
	"context"

	"github.com/cenkalti/backoff/v5"
	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
	"github.com/iPatrushevSergey/metrics/internal/retry"
)

// DefaultRetryConfig returns the default configuration: 3 repetitions with intervals of 1s, 3s, 5s
func DefaultRetryConfig() retry.RetryConfig {
	return retry.DefaultRetryConfig()
}

// RetryRepository a decorator for adding retry logic to a repository
type RetryRepository struct {
	repo       repository.MetricRepository
	config     retry.RetryConfig
	classifier *PostgresErrorClassifier
}

// NewRetryRepository creates a new repository with retry logic
func NewRetryRepository(repo repository.MetricRepository, config retry.RetryConfig) repository.MetricRepository {
	return &RetryRepository{
		repo:       repo,
		config:     config,
		classifier: NewPostgresErrorClassifier(),
	}
}

// retry performs an operation with retry logic for retriable errors
func (r *RetryRepository) retry(ctx context.Context, operation func() error) error {
	backoffStrategy := retry.NewFixedIntervalBackoff(r.config)

	_, err := backoff.Retry(ctx, func() (struct{}, error) {
		err := operation()
		if err == nil {
			return struct{}{}, nil
		}

		if !r.classifier.IsRetriable(err) {
			// An error that cannot be repeated - we stop trying
			return struct{}{}, backoff.Permanent(err)
		}

		// Retriable error - we return it for a retry
		return struct{}{}, err
	}, backoff.WithBackOff(backoffStrategy), backoff.WithMaxTries(r.config.MaxRetries+1))

	return err
}

func (r *RetryRepository) GetByID(ctx context.Context, id string) (model.Metric, error) {
	var result model.Metric
	var resultErr error

	err := r.retry(ctx, func() error {
		result, resultErr = r.repo.GetByID(ctx, id)
		return resultErr
	})

	if err != nil {
		return model.Metric{}, err
	}
	return result, nil
}

func (r *RetryRepository) GetByIDs(ctx context.Context, ids []string) (map[string]model.Metric, error) {
	var result map[string]model.Metric
	var resultErr error

	err := r.retry(ctx, func() error {
		result, resultErr = r.repo.GetByIDs(ctx, ids)
		return resultErr
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RetryRepository) GetAll(ctx context.Context) (map[string]model.Metric, error) {
	var result map[string]model.Metric
	var resultErr error

	err := r.retry(ctx, func() error {
		result, resultErr = r.repo.GetAll(ctx)
		return resultErr
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RetryRepository) Create(ctx context.Context, metric model.Metric) error {
	return r.retry(ctx, func() error {
		return r.repo.Create(ctx, metric)
	})
}

func (r *RetryRepository) CreateBatch(ctx context.Context, metrics []model.Metric) error {
	return r.retry(ctx, func() error {
		return r.repo.CreateBatch(ctx, metrics)
	})
}

func (r *RetryRepository) Update(ctx context.Context, id string, metric model.Metric) error {
	return r.retry(ctx, func() error {
		return r.repo.Update(ctx, id, metric)
	})
}

func (r *RetryRepository) UpdateBatch(ctx context.Context, metrics []model.Metric) error {
	return r.retry(ctx, func() error {
		return r.repo.UpdateBatch(ctx, metrics)
	})
}

func (r *RetryRepository) Ping(ctx context.Context) error {
	return r.retry(ctx, func() error {
		return r.repo.Ping(ctx)
	})
}
