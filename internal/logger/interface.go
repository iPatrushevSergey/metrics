package logger

import "go.uber.org/zap"

// Logger is the logging interface used by the application.
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}

// ZapLoggerAdapter adapts zap.Logger to the Logger interface.
type ZapLoggerAdapter struct {
	logger *zap.Logger
}

// NewZapLoggerAdapter returns a Logger implementation backed by the given zap.Logger.
func NewZapLoggerAdapter(logger *zap.Logger) *ZapLoggerAdapter {
	return &ZapLoggerAdapter{logger: logger}
}

func (a *ZapLoggerAdapter) Debug(msg string, fields ...zap.Field) {
	a.logger.Debug(msg, fields...)
}

func (a *ZapLoggerAdapter) Info(msg string, fields ...zap.Field) {
	a.logger.Info(msg, fields...)
}

func (a *ZapLoggerAdapter) Warn(msg string, fields ...zap.Field) {
	a.logger.Warn(msg, fields...)
}

func (a *ZapLoggerAdapter) Error(msg string, fields ...zap.Field) {
	a.logger.Error(msg, fields...)
}

func (a *ZapLoggerAdapter) Fatal(msg string, fields ...zap.Field) {
	a.logger.Fatal(msg, fields...)
}
