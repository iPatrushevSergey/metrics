package converter

import (
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/stretchr/testify/assert"
)

func TestAuditConverter_toModel(t *testing.T) {
	c := AuditConverterImpl{}
	event := dto.AuditEvent{TS: 1, Metrics: []string{"a"}, IPAddress: "127.0.0.1"}
	row := c.ToModel(event)
	assert.Equal(t, int64(1), row.TS)
	assert.Equal(t, event.Metrics, row.Metrics)
	assert.Equal(t, "127.0.0.1", row.IPAddress)
}
