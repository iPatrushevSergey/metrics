package audit

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
)

var errPublisherClosed = errors.New("audit event publisher closed")

// AuditEventPublisher fans out events to subscriber channels.
type AuditEventPublisher struct {
	log     port.Logger
	subSize int
	mu      sync.RWMutex
	subs    map[string]chan dto.AuditEvent
	closed  bool
}

// NewAuditEventPublisher creates a publisher with no subscribers.
func NewAuditEventPublisher(log port.Logger, subSize int) *AuditEventPublisher {
	return &AuditEventPublisher{
		log:     log,
		subSize: subSize,
		subs:    make(map[string]chan dto.AuditEvent),
	}
}

// Publish delivers a copy of the event to every subscriber channel.
func (p *AuditEventPublisher) Publish(e dto.AuditEvent) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed || len(p.subs) == 0 {
		return
	}

	for id, ch := range p.subs {
		select {
		case ch <- cloneEvent(e):
		default:
			p.log.Warn("audit event dropped", "subscriber", id, "reason", "queue is full")
		}
	}
}

// Subscribe registers a subscriber and returns its event channel.
func (p *AuditEventPublisher) Subscribe(subscriberID string) (<-chan dto.AuditEvent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, errPublisherClosed
	}
	if subscriberID == "" {
		return nil, fmt.Errorf("subscriber id is required")
	}
	if _, exists := p.subs[subscriberID]; exists {
		return nil, fmt.Errorf("subscriber %q already exists", subscriberID)
	}

	ch := make(chan dto.AuditEvent, p.subSize)
	p.subs[subscriberID] = ch
	return ch, nil
}

// Unsubscribe removes a subscriber and closes its channel.
func (p *AuditEventPublisher) Unsubscribe(subscriberID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch, ok := p.subs[subscriberID]
	if !ok {
		return
	}
	close(ch)
	delete(p.subs, subscriberID)
}

// Close closes all subscriber channels.
func (p *AuditEventPublisher) Close(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true
	for id, ch := range p.subs {
		close(ch)
		delete(p.subs, id)
	}
	return nil
}

func cloneEvent(e dto.AuditEvent) dto.AuditEvent {
	return dto.AuditEvent{
		TS:        e.TS,
		IPAddress: e.IPAddress,
		Metrics:   slices.Clone(e.Metrics),
	}
}

var _ port.AuditPublisher = (*AuditEventPublisher)(nil)
