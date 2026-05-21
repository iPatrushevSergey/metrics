package entity

import "errors"

var (
	// ErrMissingCounterDelta indicates a counter metric without delta.
	ErrMissingCounterDelta = errors.New("counter requires delta")

	// ErrMissingGaugeValue indicates a gauge metric without value.
	ErrMissingGaugeValue = errors.New("gauge requires value")

	// ErrUnsupportedMetricType indicates metric type is not counter or gauge.
	ErrUnsupportedMetricType = errors.New("unsupported metric type")

	// ErrMetricTypeMismatch indicates two metrics have different types.
	ErrMetricTypeMismatch = errors.New("metric type mismatch")
)
