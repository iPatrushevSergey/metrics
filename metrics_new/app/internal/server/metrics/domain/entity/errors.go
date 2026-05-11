package entity

import "errors"

var (
	// ErrMissingCounterDelta indicates a counter metric without delta.
	ErrMissingCounterDelta = errors.New("counter requires delta")

	// ErrMissingGaugeValue indicates a gauge metric without value.
	ErrMissingGaugeValue = errors.New("gauge requires value")

	// ErrUnsupportedMetricType indicates metric type is not counter or gauge.
	ErrUnsupportedMetricType = errors.New("unsupported metric type")

	// ErrMetricIdentityMismatch indicates stored and incoming metrics differ by id or MType.
	ErrMetricIdentityMismatch = errors.New("metric identity mismatch")
)
