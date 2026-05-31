package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/metrics/converter"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/metrics/model"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
)

// MetricFileRepository persists metrics to a file.
type MetricFileRepository struct {
	filePath string
	conv     converter.MetricConverter
}

// NewMetricFileRepository creates a file repository for the given path.
func NewMetricFileRepository(filePath string) *MetricFileRepository {
	return &MetricFileRepository{
		filePath: filePath,
		conv:     &converter.MetricConverterImpl{},
	}
}

// SaveAll overwrites the file with the current metrics snapshot.
func (r *MetricFileRepository) SaveAll(_ context.Context, metrics map[string]entity.Metric) error {
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return err
	}
	defer file.Close()

	rows := make([]model.Metric, 0, len(metrics))
	for _, m := range metrics {
		rows = append(rows, r.conv.ToModel(m))
	}

	return json.NewEncoder(file).Encode(rows)
}

// LoadAll reads metrics from the file.
func (r *MetricFileRepository) LoadAll(_ context.Context) ([]entity.Metric, error) {
	file, err := os.OpenFile(r.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() == 0 {
		return nil, nil
	}

	var rows []model.Metric
	if err := json.NewDecoder(file).Decode(&rows); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, err
	}

	metrics := make([]entity.Metric, 0, len(rows))
	for _, row := range rows {
		metrics = append(metrics, r.conv.ToEntity(row))
	}
	return metrics, nil
}
