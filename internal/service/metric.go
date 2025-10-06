package service

import (
	"github.com/google/uuid"
	"github.com/iPatrushevSergey/metrics/internal/model"
)

type MetricsService struct {
	metricRepo MetricRepository
}

func NewMetricService(repo MetricRepository) *MetricsService {
	return &MetricsService{metricRepo: repo}
}

func (s *MetricsService) Update(mType, mName string, value any) {
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
		}
		s.metricRepo.Create(mName, metric)
		return
	}

	switch v := value.(type) {
	case float64:
		metric.Value = &v
	case int64:
		*metric.Delta += v
	}
	s.metricRepo.Update(mName, metric)
}
