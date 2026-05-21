package application

import "errors"

// Application-level errors returned to presentation.
var (
	ErrNotFound       = errors.New("metric not found")
	ErrAlreadyExists  = errors.New("metric already exists")
	ErrBadMetricType  = errors.New("invalid metric type")
	ErrBadMetricValue = errors.New("invalid metric value")
	ErrInternal       = errors.New("internal error")
)
