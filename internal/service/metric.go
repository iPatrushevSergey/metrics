package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
)

type MetricsService struct {
	metricRepo inmemory.MetricRepository
}

func NewMetricService(repo inmemory.MetricRepository) *MetricsService {
	return &MetricsService{metricRepo: repo}
}

func (s *MetricsService) Get(mName string) (model.Metric, error) {
	metric, exists := s.metricRepo.GetByName(mName)
	if !exists {
		return model.Metric{}, fmt.Errorf("metric not found")
	}
	return metric, nil
}

func (s *MetricsService) GetAll() map[string]model.Metric {
	return s.metricRepo.GetAll()
}

func (s *MetricsService) Update(mType, mName string, value any) error {
	metric, exists := s.metricRepo.GetByName(mName)

	// If there is no object, I should return the error that the object was not found.
	// This implementation is similar to upsert
	if !exists {
		metric = model.Metric{
			ID:    uuid.NewString(),
			MType: mType,
		}

		switch v := value.(type) {
		case float64:
			metric.Value = &v
		case int64:
			metric.Delta = &v
		default:
			return fmt.Errorf("unsupported value type: %T", value)
		}
		s.metricRepo.Create(mName, metric)
		return nil
	}

	switch v := value.(type) {
	case float64:
		metric.Value = &v
	case int64:
		*metric.Delta += v
	default:
		return fmt.Errorf("unsupported value type: %T", value)
	}
	s.metricRepo.Update(mName, metric)
	return nil
}
