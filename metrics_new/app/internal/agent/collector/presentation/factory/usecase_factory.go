package factory

import "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"

// UseCaseFactory provides collector use cases to the presentation layer.
type UseCaseFactory interface {
	PollRuntimeTick() port.BackgroundRunner
	PollGopsutilTick() port.BackgroundRunner
	ReportBatchTick() port.BackgroundRunner
}
