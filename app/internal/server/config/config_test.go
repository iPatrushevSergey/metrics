package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseListenAddress(t *testing.T) {
	addr, err := parseListenAddress("localhost:8080")
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", addr)

	addr, err = parseListenAddress("http://127.0.0.1:9090")
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9090", addr)
}

func TestParseListenAddress_invalid(t *testing.T) {
	_, err := parseListenAddress("not-an-address")
	assert.Error(t, err)
}

func TestParseDuration(t *testing.T) {
	d, err := parseDuration("30")
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, d)

	d, err = parseDuration(5)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, d)

	d, err = parseDuration(int64(7))
	require.NoError(t, err)
	assert.Equal(t, 7*time.Second, d)
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

func withArgs(t *testing.T, args ...string) {
	t.Helper()
	old := os.Args
	os.Args = append([]string{"metrics-server"}, args...)
	t.Cleanup(func() { os.Args = old })
}

func writeServerConfig(t *testing.T, yaml string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "server.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
	return path
}

func TestLoadConfig_defaultValues(t *testing.T) {
	path := writeServerConfig(t, `
server:
  address: "127.0.0.1:8080"
  shutdown_timeout: "10s"
audit:
  audit_sub_size: 500
`)
	withArgs(t, "-c", path)
	t.Setenv("ADDRESS", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", cfg.Server.Address)
	assert.Equal(t, 500, cfg.Audit.AuditSubSize)
}

func TestLoadConfig_customYAML(t *testing.T) {
	path := writeServerConfig(t, `
logger:
  level: warn
server:
  address: "localhost:9091"
  shutdown_timeout: "3s"
  store_interval: "0s"
audit:
  audit_sub_size: 42
`)
	withArgs(t, "-c", path)
	t.Setenv("ADDRESS", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "localhost:9091", cfg.Server.Address)
	assert.Equal(t, 42, cfg.Audit.AuditSubSize)
	assert.Equal(t, "warn", cfg.Logger.Level)
}

func TestLoadConfig_envAddress(t *testing.T) {
	path := writeServerConfig(t, `
server:
  address: "127.0.0.1:8080"
  shutdown_timeout: "10s"
audit:
  audit_sub_size: 10
`)
	withArgs(t, "-c", path)
	t.Setenv("ADDRESS", "localhost:7070")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "localhost:7070", cfg.Server.Address)
}
