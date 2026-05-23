package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseListenAddress(t *testing.T) {
	addr, err := parseListenAddress("localhost:8080")
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", addr)
}

func TestParseDuration(t *testing.T) {
	d, err := parseDuration("30")
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, d)

	d, err = parseDuration(5)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, d)
}

func TestFinalizeConfig_ok(t *testing.T) {
	cfg := Config{
		Server: Server{
			Address:          "localhost:8080",
			ShutdownTimeout:  time.Second,
			StoreInterval:    0,
			AuditHTTPTimeout: 0,
		},
		Audit: Audit{AuditSubSize: 10},
	}
	require.NoError(t, finalizeConfig(&cfg))
	assert.Equal(t, "localhost:8080", cfg.Server.Address)
}

func TestFinalizeConfig_invalid(t *testing.T) {
	cfg := Config{
		Server: Server{
			Address:         "bad",
			ShutdownTimeout: time.Second,
		},
		Audit: Audit{AuditSubSize: 1},
	}
	assert.Error(t, finalizeConfig(&cfg))

	cfg2 := Config{
		Server: Server{
			Address:         "localhost:8080",
			ShutdownTimeout: 0,
		},
		Audit: Audit{AuditSubSize: 1},
	}
	assert.Error(t, finalizeConfig(&cfg2))
}

func TestFinalizeConfig_invalidAuditSubSize(t *testing.T) {
	cfg := Config{
		Server: Server{Address: "localhost:8080", ShutdownTimeout: time.Second},
		Audit:  Audit{AuditSubSize: 0},
	}
	assert.Error(t, finalizeConfig(&cfg))
}

func TestParseDuration_invalid(t *testing.T) {
	_, err := parseDuration("not-a-duration")
	assert.Error(t, err)
}
