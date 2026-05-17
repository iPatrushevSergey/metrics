package factory

import "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"

// UseCaseFactory provides collector use cases to the presentation layer.
type UseCaseFactory interface {
	PollRuntimeTick() port.UseCase[struct{}, int]
	PollGopsutilTick() port.UseCase[struct{}, int]
	ReportBatchTick() port.UseCase[struct{}, int]
}
