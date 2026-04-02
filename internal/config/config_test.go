package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestAddress_Set(t *testing.T) {
	var a Address
	err := a.Set("http://localhost:8080")
	require.NoError(t, err)
	assert.Equal(t, "localhost", a.Host)
	assert.Equal(t, 8080, a.Port)
}

func TestAddress_Set_autoScheme(t *testing.T) {
	var a Address
	err := a.Set("localhost:9090")
	require.NoError(t, err)
	assert.Equal(t, "http", a.Schema)
	assert.Equal(t, 9090, a.Port)
}

func TestAddress_String(t *testing.T) {
	a := Address{Host: "h", Port: 8080}
	assert.Equal(t, "h:8080", a.String())
}

func TestAddress_URL(t *testing.T) {
	a := Address{Schema: "https", Host: "x", Port: 443}
	assert.Equal(t, "https://x:443", a.URL())
}

func TestDuration_String(t *testing.T) {
	d := Duration{Duration: 5 * time.Second}
	assert.Equal(t, "5s", d.String())
}

func TestDuration_Set(t *testing.T) {
	var d Duration
	err := d.Set("10")
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, d.Duration)

	err = d.Set("2m")
	require.NoError(t, err)
	assert.Equal(t, 2*time.Minute, d.Duration)
}

func TestAddress_UnmarshalText(t *testing.T) {
	var a Address
	err := a.UnmarshalText([]byte("http://host:90"))
	require.NoError(t, err)
	assert.Equal(t, "host", a.Host)
	assert.Equal(t, 90, a.Port)
}

func TestDuration_UnmarshalText(t *testing.T) {
	var d Duration
	err := d.UnmarshalText([]byte("30s"))
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, d.Duration)
}

func TestAgentConfig_MarshalLogObject(t *testing.T) {
	c := AgentConfig{
		Address:        "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		RateLimit:      2,
		LogLevel:       "info",
	}
	err := c.MarshalLogObject(zapcore.NewMapObjectEncoder())
	require.NoError(t, err)
}

func TestServerConfig_MarshalLogObject(t *testing.T) {
	c := ServerConfig{
		Address:          "h:8080",
		LogLevel:         "debug",
		StoreInterval:    time.Minute,
		FileStoragePath:  "/tmp/m.json",
		Restore:          true,
		DatabaseDSN:      "postgres://x",
		EnableRetry:      false,
		Key:              "k",
		AuditFilePath:    "/audit.log",
		AuditURL:         "http://a/audit",
		AuditHTTPTimeout: 3 * time.Second,
	}
	err := c.MarshalLogObject(zapcore.NewMapObjectEncoder())
	require.NoError(t, err)
}

func TestAddress_Set_emptyHost(t *testing.T) {
	var a Address
	err := a.Set("http:///")
	require.Error(t, err)
}

func TestDuration_Set_invalid(t *testing.T) {
	var d Duration
	err := d.Set("not-a-valid-duration")
	require.Error(t, err)
}
