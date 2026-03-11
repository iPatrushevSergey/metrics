package audit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPublisher_noObservers(t *testing.T) {
	p := NewPublisher(nil)
	p.Notify(Event{TS: 1, Metrics: []string{"m1"}})
	err := p.Close(context.Background())
	require.NoError(t, err)
}

func TestNewPublisher_nilObserverSkipped(t *testing.T) {
	p := NewPublisher(nil, nil)
	p.Notify(Event{})
	err := p.Close(context.Background())
	require.NoError(t, err)
}
