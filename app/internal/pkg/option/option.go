package option

// Option is a generic functional option type that can be reused
// across different constructors in the application.
type Option[T any] func(*T)

// Apply applies all options to the target.
func Apply[T any](target *T, opts ...Option[T]) {
	for _, o := range opts {
		if o != nil {
			o(target)
		}
	}
}
