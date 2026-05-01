package port

import "context"

// BackgroundRunner is the contract for background task use cases.
type BackgroundRunner interface {
	Run(ctx context.Context) (int, error)
}
