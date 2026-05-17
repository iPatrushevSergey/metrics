package worker

import (
	"context"
	"errors"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
)

// RunPoolTickerLoop runs the background use case on each tick until ctx is canceled.
func RunPoolTickerLoop(
	ctx context.Context,
	uc port.UseCase[struct{}, int],
	name string,
	log port.Logger,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info("worker started", "name", name, "interval", interval.String())
	for {
		select {
		case <-ctx.Done():
			log.Info("worker stopped", "name", name)
			return
		case <-ticker.C:
			if _, err := uc.Execute(ctx, struct{}{}); err != nil && !errors.Is(err, context.Canceled) {
				log.Error("tick failed", "name", name, "error", err)
			}
		}
	}
}
