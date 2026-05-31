package usecase

import (
	"context"
	"fmt"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
)

// PingDB checks the availability of the database.
type PingDB struct {
	metricReader port.MetricReader
}

// NewPingDB returns the DB ping use case.
func NewPingDB(metricReader port.MetricReader) port.UseCase[struct{}, struct{}] {
	return &PingDB{metricReader: metricReader}
}

// Execute checks the availability of the database.
func (uc *PingDB) Execute(ctx context.Context, _ struct{}) (struct{}, error) {
	if err := uc.metricReader.Ping(ctx); err != nil {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}
	return struct{}{}, nil
}
