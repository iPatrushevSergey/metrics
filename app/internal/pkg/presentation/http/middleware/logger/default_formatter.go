package logger

import "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"

// DefaultLogFormatter implements standard logging logic.
type DefaultLogFormatter struct{}

func (f *DefaultLogFormatter) Log(log port.Logger, p LogParams) {
	log.Info("HTTP request",
		"uri", p.Ctx.Request.RequestURI,
		"method", p.Ctx.Request.Method,
		"duration", p.Duration,
		"status", p.Ctx.Writer.Status(),
		"size", p.Ctx.Writer.Size(),
	)
	log.Debug("HTTP request/response body",
		"request_body", string(p.RequestBody),
		"response_body", p.ResponseBody.String(),
	)
}
