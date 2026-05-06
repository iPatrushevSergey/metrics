package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/config"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/logger"
)

// Run runs the agent application.
func Run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	var _ port.Logger = (*logger.ZapLogger)(nil)
	zl, err := logger.NewZapLogger(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer zl.Sync()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	app := NewAgentApp(cfg, zl)
	return app.Run(ctx)
}
