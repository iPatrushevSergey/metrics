package repository

import (
	"context"
	"errors"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

var (
	// ErrNotFound it is returned when the metric is not found in the storage.
	ErrNotFound = errors.New("metric not found")
	// ErrAlreadyExists it is returned when an attempt is made to create a metric with an existing ID.
	ErrAlreadyExists = errors.New("metric already exists")
)

// MetricRepository defines the interface for working with the metric repository
type MetricRepository interface {
	// GetByID returns a metric by ID
	GetByID(ctx context.Context, id string) (model.Metric, error)
	// GetByIDs returns metrics based on a list of IDs
	GetByIDs(ctx context.Context, ids []string) (map[string]model.Metric, error)
	// GetAll returns all metrics from storage
	GetAll(ctx context.Context) (map[string]model.Metric, error)
	// Create creates a new metric in the storage
	Create(ctx context.Context, metric model.Metric) error
	// CreateBatch creates multiple metrics in one call
	CreateBatch(ctx context.Context, metrics []model.Metric) error
	// Update updates an existing metric by ID
	Update(ctx context.Context, id string, metric model.Metric) error
	// UpdateBatch updates multiple metrics in one call
	UpdateBatch(ctx context.Context, metrics []model.Metric) error
	// Ping checks storage availability
	Ping(ctx context.Context) error
}
