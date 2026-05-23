package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withArgs(t *testing.T, args ...string) {
	t.Helper()
	old := os.Args
	os.Args = append([]string{"metrics-agent"}, args...)
	t.Cleanup(func() { os.Args = old })
}

func writeAgentConfig(t *testing.T, yaml string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "agent.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0600))
	return path
}

func TestLoadConfig_defaultValues(t *testing.T) {
	path := writeAgentConfig(t, `
agent:
  address: "http://127.0.0.1:8080"
  poll_interval: "2s"
  report_interval: "10s"
`)
	withArgs(t, "-c", path)
	t.Setenv("ADDRESS", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:8080", cfg.Agent.Address)
	assert.Equal(t, 2*time.Second, cfg.Agent.PollInterval)
}

func TestLoadConfig_customYAML(t *testing.T) {
	path := writeAgentConfig(t, `
logger:
  level: debug
agent:
  address: "http://example.com:8088"
  poll_interval: "4s"
  report_interval: "8s"
  rate_limit: 2
`)
	withArgs(t, "-c", path)
	t.Setenv("ADDRESS", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "http://example.com:8088", cfg.Agent.Address)
	assert.Equal(t, 4*time.Second, cfg.Agent.PollInterval)
	assert.Equal(t, 2, cfg.Agent.RateLimit)
}

func TestDuration_UnmarshalText(t *testing.T) {
	var d Duration
	require.NoError(t, d.UnmarshalText([]byte("10")))
	assert.Equal(t, 10*time.Second, d.Duration)
	assert.Equal(t, "10s", d.String())
}
