package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress_Set(t *testing.T) {
	var a Address
	require.NoError(t, a.Set("localhost:8080"))
	assert.Equal(t, "http", a.Schema)
	assert.Equal(t, "localhost", a.Host)
	assert.Equal(t, 8080, a.Port)
	assert.Equal(t, "http://localhost:8080", a.URL())
}

func TestAddress_Set_invalid(t *testing.T) {
	var a Address
	assert.Error(t, a.Set("localhost"))
}

func TestAddress_String(t *testing.T) {
	a := Address{Host: "localhost", Port: 8080}
	assert.Equal(t, "localhost:8080", a.String())
}

func TestDuration_Set(t *testing.T) {
	var d Duration
	require.NoError(t, d.Set("5"))
	assert.Equal(t, 5*time.Second, d.Duration)

	require.NoError(t, d.Set("2s"))
	assert.Equal(t, 2*time.Second, d.Duration)
	assert.Equal(t, "2s", d.String())
}

func TestDuration_UnmarshalText(t *testing.T) {
	var d Duration
	require.NoError(t, d.UnmarshalText([]byte("3")))
	assert.Equal(t, 3*time.Second, d.Duration)
}

func TestAddress_UnmarshalText(t *testing.T) {
	var a Address
	require.NoError(t, a.UnmarshalText([]byte("localhost:9090")))
	assert.Equal(t, 9090, a.Port)
}
