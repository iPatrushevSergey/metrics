package audit

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mailru/easyjson"
)

// HTTPObserver sends audit events to a configured HTTP.
type HTTPObserver struct {
	url    string
	client *http.Client
}

// NewHTTPObserver creates an HTTPObserver for the given URL.
func NewHTTPObserver(rawURL string, client *http.Client) (*HTTPObserver, error) {
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid audit URL: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("invalid audit URL: scheme and host are required")
	}

	return &HTTPObserver{
		url:    parsedURL.String(),
		client: client,
	}, nil
}

func (o *HTTPObserver) Publish(ctx context.Context, e Event) error {
	b, err := easyjson.Marshal(e)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("audit HTTP endpoint returned status %d", resp.StatusCode)
	}
	return nil
}

// Close implements Observer. HTTPObserver has no resources to release.
func (o *HTTPObserver) Close() error {
	return nil
}
