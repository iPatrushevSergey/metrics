package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewHTTPObserver_nilClient(t *testing.T) {
	_, err := NewHTTPObserver("http://example.com/path", nil)
	require.Error(t, err)
}

func TestNewHTTPObserver_invalidURL(t *testing.T) {
	_, err := NewHTTPObserver("://bad", http.DefaultClient)
	require.Error(t, err)
}

func TestNewHTTPObserver_missingHost(t *testing.T) {
	_, err := NewHTTPObserver("http:///only-path", http.DefaultClient)
	require.Error(t, err)
}

func TestHTTPObserver_Publish_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	obs, err := NewHTTPObserver(srv.URL, srv.Client())
	require.NoError(t, err)
	t.Cleanup(func() { _ = obs.Close() })

	require.NoError(t, obs.Publish(context.Background(), Event{TS: 1, Metrics: []string{"x"}}))
}

func TestHTTPObserver_Publish_badStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	obs, err := NewHTTPObserver(srv.URL, srv.Client())
	require.NoError(t, err)

	err = obs.Publish(context.Background(), Event{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "500")
}

func TestHTTPObserver_Close(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	obs, err := NewHTTPObserver(srv.URL, srv.Client())
	require.NoError(t, err)
	require.NoError(t, obs.Close())
}
