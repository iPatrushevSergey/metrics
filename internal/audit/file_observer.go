package audit

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/mailru/easyjson"
)

var (
	errFileObserverClosed = errors.New("audit file observer closed")
)

// FileObserver writes events to a file.
type FileObserver struct {
	file *os.File

	mu     sync.Mutex
	closed bool
}

// NewFileObserver creates a file observer.
func NewFileObserver(path string) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return &FileObserver{file: file}, nil
}

func (o *FileObserver) Publish(ctx context.Context, e Event) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.closed {
		return errFileObserverClosed
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	data, err := easyjson.Marshal(cloneEvent(e))
	if err != nil {
		return err
	}

	if _, err = o.file.Write(data); err != nil {
		return err
	}
	if _, err = o.file.Write([]byte{'\n'}); err != nil {
		return err
	}

	return nil
}

// Close closes the file observer.
func (o *FileObserver) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.closed {
		return nil
	}

	o.closed = true
	return o.file.Close()
}
