package config

import (
	"os"
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

func TestLoadAgentConfig_fileFlagEnvPriority(t *testing.T) {
	resetConfigEnv(t)

	configPath := writeTempConfigFile(t, `{
		"address": "filehost:8081",
		"poll_interval": "4s",
		"report_interval": "7s",
		"crypto_key": "file-public.pem",
		"key": "file-key",
		"rate_limit": 9,
		"log_level": "debug"
	}`)

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{
		"agent",
		"-config", configPath,
		"-r", "11s",
		"-crypto-key", "flag-public.pem",
	}

	t.Setenv("ADDRESS", "envhost:9090")
	t.Setenv("POLL_INTERVAL", "13s")

	cfg, err := LoadAgentConfig()
	require.NoError(t, err)

	require.Equal(t, "http://envhost:9090", cfg.Address)
	require.Equal(t, 13*time.Second, cfg.PollInterval)
	require.Equal(t, 11*time.Second, cfg.ReportInterval)
	require.Equal(t, "flag-public.pem", cfg.CryptoKey)
	require.Equal(t, "file-key", cfg.Key)
	require.Equal(t, 9, cfg.RateLimit)
	require.Equal(t, "debug", cfg.LogLevel)
}

func TestLoadAgentConfig_explicitJSONValuesOverrideDefaults(t *testing.T) {
	resetConfigEnv(t)

	configPath := writeTempConfigFile(t, `{
		"key": "json-key",
		"log_level": ""
	}`)

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"agent", "-config", configPath}

	cfg, err := LoadAgentConfig()
	require.NoError(t, err)

	require.Equal(t, "http://127.0.0.1:8080", cfg.Address)
	require.Equal(t, 2*time.Second, cfg.PollInterval)
	require.Equal(t, 10*time.Second, cfg.ReportInterval)
	require.Equal(t, "", cfg.CryptoKey)
	require.Equal(t, "json-key", cfg.Key)
	require.Equal(t, 0, cfg.RateLimit)
	require.Equal(t, "", cfg.LogLevel)
}

func TestLoadServerConfig_fileFlagEnvPriority(t *testing.T) {
	resetConfigEnv(t)

	configPath := writeTempConfigFile(t, `{
		"address": "filehost:8081",
		"restore": false,
		"store_interval": "4s",
		"store_file": "file-store.json",
		"database_dsn": "postgres://file",
		"crypto_key": "file-private.pem",
		"key": "file-key",
		"enable_retry": false,
		"log_level": "debug",
		"audit_file": "file-audit.log",
		"audit_url": "http://file/audit",
		"audit_http_timeout": "6s"
	}`)

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{
		"server",
		"-c", configPath,
		"-f", "flag-store.json",
		"-crypto-key", "flag-private.pem",
	}

	t.Setenv("ADDRESS", "envhost:9090")
	t.Setenv("RESTORE", "true")
	t.Setenv("STORE_INTERVAL", "15s")

	cfg, err := LoadServerConfig()
	require.NoError(t, err)

	require.Equal(t, "envhost:9090", cfg.Address)
	require.True(t, cfg.Restore)
	require.Equal(t, 15*time.Second, cfg.StoreInterval)
	require.Equal(t, "flag-store.json", cfg.FileStoragePath)
	require.Equal(t, "postgres://file", cfg.DatabaseDSN)
	require.Equal(t, "flag-private.pem", cfg.CryptoKey)
	require.Equal(t, "file-key", cfg.Key)
	require.False(t, cfg.EnableRetry)
	require.Equal(t, "debug", cfg.LogLevel)
	require.Equal(t, "file-audit.log", cfg.AuditFilePath)
	require.Equal(t, "http://file/audit", cfg.AuditURL)
	require.Equal(t, 6*time.Second, cfg.AuditHTTPTimeout)
}

func TestLoadServerConfig_defaults(t *testing.T) {
	resetConfigEnv(t)

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"server"}

	cfg, err := LoadServerConfig()
	require.NoError(t, err)

	require.Equal(t, "127.0.0.1:8080", cfg.Address)
	require.Equal(t, "info", cfg.LogLevel)
	require.Equal(t, 300*time.Second, cfg.StoreInterval)
	require.Equal(t, "metrics.json", cfg.FileStoragePath)
	require.True(t, cfg.Restore)
	require.True(t, cfg.EnableRetry)
	require.Equal(t, 2*time.Second, cfg.AuditHTTPTimeout)
}

func writeTempConfigFile(t *testing.T, body string) string {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "config-*.json")
	require.NoError(t, err)
	_, err = f.WriteString(body)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	return f.Name()
}

func resetConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"ADDRESS",
		"POLL_INTERVAL",
		"REPORT_INTERVAL",
		"KEY",
		"CRYPTO_KEY",
		"CONFIG",
		"RATE_LIMIT",
		"LOG_LEVEL",
		"STORE_INTERVAL",
		"FILE_STORAGE_PATH",
		"RESTORE",
		"DATABASE_DSN",
		"ENABLE_RETRY",
		"AUDIT_FILE",
		"AUDIT_URL",
		"AUDIT_HTTP_TIMEOUT",
	} {
		t.Setenv(key, "")
	}
}
