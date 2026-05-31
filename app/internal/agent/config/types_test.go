package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress_Set(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSchema string
		wantHost   string
		wantPort   int
		wantErr    bool
	}{
		{"host:port", "localhost:8080", "http", "localhost", 8080, false},
		{"with http", "http://example.com:443", "http", "example.com", 443, false},
		{"no port", "localhost", "", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Address
			err := a.Set(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSchema, a.Schema)
			assert.Equal(t, tt.wantHost, a.Host)
			assert.Equal(t, tt.wantPort, a.Port)
		})
	}
}

func TestAddress_URL(t *testing.T) {
	a := Address{Schema: "http", Host: "localhost", Port: 8080}
	assert.Equal(t, "http://localhost:8080", a.URL())
}

func TestAddress_String(t *testing.T) {
	a := Address{Host: "localhost", Port: 8080}
	assert.Equal(t, "localhost:8080", a.String())
}

func TestDuration_Set(t *testing.T) {
	t.Run("integer seconds", func(t *testing.T) {
		var d Duration
		require.NoError(t, d.Set("60"))
		assert.Equal(t, 60*time.Second, d.Duration)
	})

	t.Run("duration string", func(t *testing.T) {
		var d Duration
		require.NoError(t, d.Set("2s"))
		assert.Equal(t, 2*time.Second, d.Duration)
	})

	t.Run("invalid", func(t *testing.T) {
		var d Duration
		assert.Error(t, d.Set("abc"))
	})
}

func TestParseURLAddress(t *testing.T) {
	url, err := parseURLAddress("127.0.0.1:9090")
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:9090", url)
}
