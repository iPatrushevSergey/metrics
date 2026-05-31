// Package port defines the presentation-layer contract for collector use cases.
package port

import "github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"

// UseCaseFactory provides collector use cases to background workers.
type UseCaseFactory interface {
	// PollRuntimeTick samples runtime.MemStats into the local metrics store.
	PollRuntimeTick() port.UseCase[struct{}, int]
	// PollGopsutilTick samples gopsutil metrics into the local metrics store.
	PollGopsutilTick() port.UseCase[struct{}, int]
	// ReportBatchTick sends accumulated metrics to the server in one batch.
	ReportBatchTick() port.UseCase[struct{}, int]
}
