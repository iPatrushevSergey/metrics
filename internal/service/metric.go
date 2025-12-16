package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
)

var (
	// ErrNotFound it is returned when the metric is not found
	ErrNotFound = errors.New("metric not found")
	// ErrBadMetricValue returned when the metric value is invalid
	ErrBadMetricValue = errors.New("invalid metric value")
	// ErrBadMetricType returned when the metric type is invalid
	ErrBadMetricType = errors.New("invalid metric type")
	// ErrInternal returned in case of internal errors in the service
	ErrInternal = errors.New("internal service error")
)

func validateMetricType(mType string) error {
	switch mType {
	case model.Gauge, model.Counter:
		return nil
	default:
		return ErrBadMetricType
	}
}

// FormatMetric formats the metric into a string representation
func (s *MetricsService) FormatMetric(metric model.Metric) (string, error) {
	return formatMetricToStr(metric)
}

func formatMetricToStr(metric model.Metric) (string, error) {
	switch metric.MType {
	case model.Gauge:
		if metric.Value == nil {
			return "", fmt.Errorf("%w: gauge value is nil", ErrInternal)
		}
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64), nil
	case model.Counter:
		if metric.Delta == nil {
			return "", fmt.Errorf("%w: counter value is nil", ErrInternal)
		}
		return strconv.FormatInt(*metric.Delta, 10), nil
	default:
		return "", fmt.Errorf("%w: unknown metric MType: %s", ErrInternal, metric.MType)
	}
}

type MetricsService struct {
	metricRepo repository.MetricRepository
}

func NewMetricService(
	repo repository.MetricRepository,
) *MetricsService {
	return &MetricsService{metricRepo: repo}
}

func (s *MetricsService) GetValue(ctx context.Context, mType, mName string) (string, error) {
	if err := validateMetricType(mType); err != nil {
		return "", err
	}

	metric, err := s.metricRepo.GetByID(ctx, mName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}

	if metric.MType != mType {
		return "", ErrNotFound
	}

	formattedMetric, err := formatMetricToStr(metric)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return formattedMetric, nil
}

// GetMetric returns a metric by type and name
func (s *MetricsService) GetMetric(ctx context.Context, mType, mName string) (model.Metric, error) {
	if err := validateMetricType(mType); err != nil {
		return model.Metric{}, err
	}

	metric, err := s.metricRepo.GetByID(ctx, mName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Metric{}, ErrNotFound
		}
		return model.Metric{}, err
	}

	if metric.MType != mType {
		return model.Metric{}, ErrNotFound
	}

	if metric.MType == model.Gauge && metric.Value == nil {
		return model.Metric{}, fmt.Errorf("%w: gauge value is nil", ErrInternal)
	}
	if metric.MType == model.Counter && metric.Delta == nil {
		return model.Metric{}, fmt.Errorf("%w: counter value is nil", ErrInternal)
	}

	return metric, nil
}

// GetAll returns all metrics from storage
func (s *MetricsService) GetAll(ctx context.Context) (map[string]model.Metric, error) {
	metrics, err := s.metricRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// Update updates or creates a metric by type, name, and value
func (s *MetricsService) Update(ctx context.Context, mType, mName string, value string) error {
	if err := validateMetricType(mType); err != nil {
		return err
	}

	metric := model.Metric{
		ID:    mName,
		MType: mType,
	}

	switch mType {
	case model.Gauge:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrBadMetricValue
		}
		metric.Value = &v
	case model.Counter:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrBadMetricValue
		}
		metric.Delta = &v
	}

	existing, err := s.metricRepo.GetByID(ctx, mName)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("%w: failed to get metric: %w", ErrInternal, err)
	}

	if errors.Is(err, repository.ErrNotFound) {
		if err := s.metricRepo.Create(ctx, metric); err != nil {
			return fmt.Errorf("%w: failed to create metric: %w", ErrInternal, err)
		}
		return nil
	}

	switch mType {
	case model.Counter:
		if metric.Delta != nil {
			if existing.Delta != nil {
				newDelta := *existing.Delta + *metric.Delta
				existing.Delta = &newDelta
			} else {
				existing.Delta = metric.Delta
			}
		}
	case model.Gauge:
		existing.Value = metric.Value
	}

	if err := s.metricRepo.Update(ctx, mName, existing); err != nil {
		return fmt.Errorf("%w: failed to update metric: %w", ErrInternal, err)
	}
	return nil
}

// UpdateJSON updates or creates a metric from the domain model
func (s *MetricsService) UpdateJSON(ctx context.Context, metric model.Metric) error {
	if err := validateMetricType(metric.MType); err != nil {
		return err
	}

	if metric.MType == model.Counter && metric.Delta == nil {
		return ErrBadMetricValue
	}
	if metric.MType == model.Gauge && metric.Value == nil {
		return ErrBadMetricValue
	}

	existing, err := s.metricRepo.GetByID(ctx, metric.ID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("%w: failed to get metric: %w", ErrInternal, err)
	}

	if errors.Is(err, repository.ErrNotFound) {
		if err := s.metricRepo.Create(ctx, metric); err != nil {
			return fmt.Errorf("%w: failed to create metric: %w", ErrInternal, err)
		}
		return nil
	}

	switch metric.MType {
	case model.Counter:
		if metric.Delta != nil {
			if existing.Delta != nil {
				newDelta := *existing.Delta + *metric.Delta
				existing.Delta = &newDelta
			} else {
				existing.Delta = metric.Delta
			}
		}
	case model.Gauge:
		existing.Value = metric.Value
	}

	if err := s.metricRepo.Update(ctx, metric.ID, existing); err != nil {
		return fmt.Errorf("%w: failed to update metric: %w", ErrInternal, err)
	}
	return nil
}

// UpdatesJSON updates or creates metrics from the domain model
func (s *MetricsService) UpdatesJSON(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Combining duplicates. For counter, we sum the deltas,
	// for gauge we overwrite the last value.
	mergedByID := make(map[string]model.Metric, len(metrics))
	for _, metric := range metrics {
		if err := validateMetricType(metric.MType); err != nil {
			return err
		}

		if metric.MType == model.Counter && metric.Delta == nil {
			return ErrBadMetricValue
		}
		if metric.MType == model.Gauge && metric.Value == nil {
			return ErrBadMetricValue
		}

		existing, exists := mergedByID[metric.ID]
		if !exists {
			mergedByID[metric.ID] = metric
			continue
		}

		// Merge with previously seen metric with the same ID
		switch metric.MType {
		case model.Counter:
			if metric.Delta != nil {
				var current int64
				if existing.Delta != nil {
					current = *existing.Delta
				}
				newDelta := current + *metric.Delta
				existing.Delta = &newDelta
			}
		case model.Gauge:
			existing.Value = metric.Value
		}

		mergedByID[metric.ID] = existing
	}

	uniqueMetrics := make([]model.Metric, 0, len(mergedByID))
	for _, m := range mergedByID {
		uniqueMetrics = append(uniqueMetrics, m)
	}

	createMetrics := make([]model.Metric, 0, len(uniqueMetrics))
	updateMetrics := make([]model.Metric, 0, len(uniqueMetrics))

	ids := make([]string, 0, len(mergedByID))
	for id := range mergedByID {
		ids = append(ids, id)
	}

	existingMetrics, err := s.metricRepo.GetByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("%w: failed to get metrics batch: %w", ErrInternal, err)
	}

	for _, metric := range uniqueMetrics {
		// Determining whether to create a new one or update an existing one
		existing, found := existingMetrics[metric.ID]
		if !found {
			createMetrics = append(createMetrics, metric)
			continue
		}

		switch metric.MType {
		case model.Counter:
			if metric.Delta != nil {
				if existing.Delta != nil {
					newDelta := *existing.Delta + *metric.Delta
					existing.Delta = &newDelta
				} else {
					existing.Delta = metric.Delta
				}
			}
		case model.Gauge:
			existing.Value = metric.Value
		}

		updateMetrics = append(updateMetrics, existing)
	}

	if len(createMetrics) > 0 {
		if err := s.metricRepo.CreateBatch(ctx, createMetrics); err != nil {
			return fmt.Errorf("%w: failed to create metrics batch: %w", ErrInternal, err)
		}
	}

	if len(updateMetrics) > 0 {
		if err := s.metricRepo.UpdateBatch(ctx, updateMetrics); err != nil {
			return fmt.Errorf("%w: failed to update metrics batch: %w", ErrInternal, err)
		}
	}

	return nil
}

// PingDB checks the availability of the data warehouse
func (s *MetricsService) PingDB(ctx context.Context) error {
	return s.metricRepo.Ping(ctx)
}
