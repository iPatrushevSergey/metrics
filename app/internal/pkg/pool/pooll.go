// Package pool provides a generic thread-safe object pool
// for types that implement Reset().
package pool

import "sync"

// Resetter is the constraint for pool-managed types.
type Resetter interface {
	Reset()
}

// Pool is a thread-safe container for reusable objects.
// On Put the object is automatically reset before being returned to the pool.
type Pool[T Resetter] struct {
	p sync.Pool
}

// New creates a Pool. The factory function is called when the pool is empty.
func New[T Resetter](factory func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any { return factory() },
		},
	}
}

// Get retrieves an object from the pool or creates a new one via factory.
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put resets v and places it back into the pool for reuse.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.p.Put(v)
}
