package dto

// GetMetricValueInput is the input for the GetMetricValue use case.
type GetMetricValueInput struct {
	MType string
	ID    string
}

// GetMetricInput loads a metric.
type GetMetricInput struct {
	MType string
	ID    string
}

// MetricOutput is a read model for internal use.
type MetricOutput struct {
	ID    string
	MType string
	Delta *int64
	Value *float64
	Hash  string
}

// MetricForDisplayOutput is one metric for presentation.
type MetricForDisplayOutput struct {
	MetricID    string
	MetricValue string
}

// UpdateMetricInput is the input for metric update.
type UpdateMetricInput struct {
	MType     string
	ID        string
	Value     string
	IPAddress string
}

// UpsertMetricInput is a single metric upsert.
type UpsertMetricInput struct {
	ID        string
	MType     string
	Delta     *int64
	Value     *float64
	Hash      string
	IPAddress string
}

// UpsertMetricsBatchInput is a batch upsert.
type UpsertMetricsBatchInput struct {
	Metrics   []UpsertMetricInput
	IPAddress string
}
