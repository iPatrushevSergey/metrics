package worker

import (
	"context"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/presentation/factory"
)

type stubUseCase struct {
	onExecute func(context.Context, struct{}) (int, error)
}

func (s stubUseCase) Execute(ctx context.Context, in struct{}) (int, error) {
	if s.onExecute != nil {
		return s.onExecute(ctx, in)
	}
	return 1, nil
}

type stubUseCaseFactory struct {
	pollRuntime  port.UseCase[struct{}, int]
	pollGopsutil port.UseCase[struct{}, int]
	reportBatch  port.UseCase[struct{}, int]
}

func (f stubUseCaseFactory) PollRuntimeTick() port.UseCase[struct{}, int] {
	return f.pollRuntime
}

func (f stubUseCaseFactory) PollGopsutilTick() port.UseCase[struct{}, int] {
	return f.pollGopsutil
}

func (f stubUseCaseFactory) ReportBatchTick() port.UseCase[struct{}, int] {
	return f.reportBatch
}

var _ factory.UseCaseFactory = stubUseCaseFactory{}
