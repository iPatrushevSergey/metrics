package converter

import (
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/metrics/model"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:file metric_gen.go
// goverter:output:package converter
type MetricConverter interface {
	ToEntity(source model.Metric) entity.Metric
	ToModel(source entity.Metric) model.Metric
}
