// Package service holds metric helpers outside entity state mutations (compare balance MetricService ↔ entity methods).
package service

import (
	"fmt"
	"strings"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/entity"
)

// MetricService performs operations on metrics.
type MetricService struct{}

// CheckMetricType normalizes and validates an external metric type label.
func (MetricService) CheckMetricType(s string) (entity.MetricType, error) {
	mType := entity.MetricType(strings.ToLower(strings.TrimSpace(s)))
	switch mType {
	case entity.Counter, entity.Gauge:
		return mType, nil
	default:
		return "", fmt.Errorf("unknown metric type")
	}
}
