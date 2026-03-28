package audit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type countingObserver struct {
	ch  chan struct{}
	err error
}

func (c *countingObserver) Publish(ctx context.Context, e Event) error {
	select {
	case c.ch <- struct{}{}:
	default:
	}
	return c.err
}

func (c *countingObserver) Close() error { return nil }

func TestPublisher_Notify_deliversToObserver(t *testing.T) {
	ch := make(chan struct{}, 1)
	obs := &countingObserver{ch: ch}
	p := NewPublisher(nil, obs)

	p.Notify(Event{TS: 7, Metrics: []string{"cpu"}})

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("observer did not receive event")
	}

	require.NoError(t, p.Close(context.Background()))
}

func TestPublisher_Close_idempotent(t *testing.T) {
	p := NewPublisher(nil)
	require.NoError(t, p.Close(context.Background()))
	require.NoError(t, p.Close(context.Background()))
}
