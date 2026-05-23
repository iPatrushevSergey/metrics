package audit

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditEventPublisher_subscribeAndPublish(t *testing.T) {
	p := NewAuditEventPublisher(logger.NewNopLogger(), 2)

	ch, err := p.Subscribe("file")
	require.NoError(t, err)

	p.Publish(dto.AuditEvent{TS: 1, Metrics: []string{"cpu"}})

	select {
	case e := <-ch:
		assert.Equal(t, int64(1), e.TS)
	default:
		t.Fatal("expected event")
	}

	p.Unsubscribe("file")
	_, ok := <-ch
	assert.False(t, ok)
}

func TestAuditEventPublisher_subscribeErrors(t *testing.T) {
	p := NewAuditEventPublisher(logger.NewNopLogger(), 1)
	_, err := p.Subscribe("")
	assert.Error(t, err)

	_, err = p.Subscribe("a")
	require.NoError(t, err)
	_, err = p.Subscribe("a")
	assert.Error(t, err)
}

func TestAuditEventPublisher_close(t *testing.T) {
	p := NewAuditEventPublisher(logger.NewNopLogger(), 1)
	ch, err := p.Subscribe("a")
	require.NoError(t, err)
	require.NoError(t, p.Close(context.Background()))
	_, ok := <-ch
	assert.False(t, ok)
}
