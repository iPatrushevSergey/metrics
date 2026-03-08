package audit

import "context"

// Observer delivers audit events to a destination.
type Observer interface {
	Publish(ctx context.Context, e Event) error
	Close() error
}

// Publisher forwards events to all registered observers.
type Publisher interface {
	Notify(e Event)
	Close(ctx context.Context) error
}
