package service

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrNotFound       = errors.New("metric not found")
	ErrBadMetricValue = errors.New("invalid metric value")
	ErrBadMetricType  = errors.New("invalid metric type")
	ErrInternal       = errors.New("internal service error")
)

type templateData struct {
	Name  string
	Value string
}

type responseMetrics struct {
	Metrics []templateData
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
	return &MetricsService{
		metricRepo: repo,
	}
}

func (s *MetricsService) GetValue(mType, mName string) (string, error) {
	mType = strings.ToLower(mType)
	mName = strings.TrimSpace(strings.ToLower(mName))

	switch mType {
	case model.Gauge, model.Counter:
	default:
		return "", ErrBadMetricType
	}

	metric, exists := s.metricRepo.GetByID(mName)

	if !exists {
		return "", ErrNotFound
	}

	if metric.MType != mType {
		return "", ErrNotFound
	}

	formattedMetric, err := formatMetricToStr(metric)
	if err != nil {
		return "", err
	}

	return formattedMetric, nil
}

func (s *MetricsService) GetMetric(mType, mName string) (model.Metric, error) {
	mType = strings.ToLower(mType)
	mName = strings.TrimSpace(strings.ToLower(mName))

	switch mType {
	case model.Gauge, model.Counter:
	default:
		return model.Metric{}, ErrBadMetricType
	}

	metric, exists := s.metricRepo.GetByID(mName)

	if !exists {
		return model.Metric{}, ErrNotFound
	}

	if metric.MType != mType {
		return model.Metric{}, ErrNotFound
	}

	switch metric.MType {
	case model.Gauge:
		if metric.Value == nil {
			return model.Metric{}, fmt.Errorf("%w: gauge value is nil", ErrInternal)
		}
	case model.Counter:
		if metric.Delta == nil {
			return model.Metric{}, fmt.Errorf("%w: counter value is nil", ErrInternal)
		}
	}

	return metric, nil
}

func (s *MetricsService) GetAll() (responseMetrics, error) {
	metrics := s.metricRepo.GetAll()

	keys := make([]string, 0, len(metrics))

	for key := range metrics {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	data := responseMetrics{}

	for _, key := range keys {
		value, err := formatMetricToStr(metrics[key])
		if err != nil {
			logger.Log.Error("error formatting metric", zap.String("key", key), zap.Error(err))
			continue
		}
		data.Metrics = append(data.Metrics, templateData{Name: key, Value: value})
	}

	return data, nil
}

func (s *MetricsService) Update(mType, mName string, value string) error {
	mType = strings.ToLower(mType)
	mName = strings.TrimSpace(strings.ToLower(mName))

	var parsedValue any
	var err error

	switch mType {
	case model.Gauge:
		parsedValue, err = strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrBadMetricValue
		}
	case model.Counter:
		parsedValue, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrBadMetricValue
		}
	default:
		return ErrBadMetricType
	}

	metric, exists := s.metricRepo.GetByID(mName)

	// If there is no object, I should return the error that the object was not found.
	// This implementation is similar to upsert
	if !exists {
		metric = model.Metric{
			ID:    mName,
			MType: mType,
		}

		switch v := parsedValue.(type) {
		case float64:
			metric.Value = &v
		case int64:
			metric.Delta = &v
		}
		s.metricRepo.Create(metric)
		return nil
	}

	switch v := parsedValue.(type) {
	case float64:
		metric.Value = &v
	case int64:
		*metric.Delta += v
	}
	s.metricRepo.Update(mName, metric)
	return nil
}

func (s *MetricsService) UpdateJSON(metric model.Metric) error {
	metric.MType = strings.ToLower(metric.MType)
	metric.ID = strings.TrimSpace(strings.ToLower(metric.ID))

	metricDB, exists := s.metricRepo.GetByID(metric.ID)

	// If there is no object, I should return the error that the object was not found.
	// This implementation is similar to upsert
	if !exists {
		s.metricRepo.Create(metric)
		return nil
	}

	switch metric.MType {
	case model.Counter:
		*metricDB.Delta += *metric.Delta
	case model.Gauge:
		metricDB.Value = metric.Value
	}

	s.metricRepo.Update(metricDB.ID, metricDB)
	return nil
}
