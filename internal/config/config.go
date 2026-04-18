// Package config provides application configuration (flags, environment, server and agent settings).
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap/zapcore"
)

// === Custom Flag Type ===

// Address - this is a custom flag type for the address 'host:port'.
type Address struct {
	Schema string
	Host   string
	Port   int
}

// String implements the interface flag.Value
func (a *Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// Set implements the interface flag.Value
func (a *Address) Set(s string) error {
	if !(strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")) {
		s = "http://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	if u.Host == "" {
		return errors.New("host is empty")
	}

	hostName, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	a.Schema = u.Scheme
	a.Host = hostName
	a.Port = port

	return nil
}

func (a *Address) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

func (a *Address) URL() string {
	return fmt.Sprintf("%s://%s:%d", a.Schema, a.Host, a.Port)
}

// Duration is a custom flag type for time duration (seconds or duration string).
type Duration struct {
	time.Duration
}

func (d *Duration) String() string {
	return d.Duration.String()
}

func (d *Duration) Set(s string) error {
	if val, err := strconv.Atoi(s); err == nil {
		d.Duration = time.Duration(val) * time.Second
		return nil
	}

	val, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = val
	return nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}

// === Agent Configuration ===

// AgentConfig - the final configuration structure for the agent.
type AgentConfig struct {
	Address        string
	PollInterval   time.Duration
	ReportInterval time.Duration
	Key            string // Key for hash calculation
	CryptoKey      string // Path to PEM file with RSA public key (optional)
	RateLimit      int
	LogLevel       string
}

func (c *AgentConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", c.Address)
	enc.AddDuration("poll_interval", c.PollInterval)
	enc.AddDuration("report_interval", c.ReportInterval)
	enc.AddInt("rate_limit", c.RateLimit)
	enc.AddString("log_level", c.LogLevel)
	return nil
}

type agentInternalConfig struct {
	Address        Address  `env:"ADDRESS" json:"address"`
	PollInterval   Duration `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval Duration `env:"REPORT_INTERVAL" json:"report_interval"`
	Key            string   `env:"KEY" json:"key"`
	CryptoKey      string   `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigPath     string   `env:"CONFIG"`
	RateLimit      int      `env:"RATE_LIMIT" json:"rate_limit"`
	LogLevel       string   `env:"LOG_LEVEL" json:"log_level"`
}

// LoadAgentConfig loads agent configuration from flags and environment (env overrides flags).
func LoadAgentConfig() (AgentConfig, error) {
	cfg := agentInternalConfig{}
	initAgentDefaults(&cfg)

	configPath, err := resolveConfigPath(os.Args[1:])
	if err != nil {
		return AgentConfig{}, fmt.Errorf("config path parsing error: %w", err)
	}
	cfg.ConfigPath = configPath
	if cfg.ConfigPath != "" {
		if err := loadJSONConfig(cfg.ConfigPath, &cfg); err != nil {
			return AgentConfig{}, fmt.Errorf("config file error: %w", err)
		}
	}

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.Var(&cfg.Address, "a", "server address")
	fs.StringVar(&cfg.ConfigPath, "c", cfg.ConfigPath, "path to JSON config file")
	fs.StringVar(&cfg.ConfigPath, "config", cfg.ConfigPath, "path to JSON config file")
	fs.Var(&cfg.ReportInterval, "r", "frequency of sending metrics (seconds or duration)")
	fs.Var(&cfg.PollInterval, "p", "frequency of metrics polling (seconds or duration)")
	fs.StringVar(&cfg.Key, "k", cfg.Key, "key for hash calculation")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path to RSA public key PEM for encrypting payloads to the server")
	fs.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "rate limit for concurrent batch requests (0 = sequential, >0 = worker pool size)")
	fs.StringVar(&cfg.LogLevel, "log", cfg.LogLevel, "logging level")

	// Flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return AgentConfig{}, fmt.Errorf("flag parsing error: %w", err)
	}

	// Env
	if err := env.Parse(&cfg); err != nil {
		return AgentConfig{}, fmt.Errorf("ENV parsing error: %w", err)
	}

	return AgentConfig{
		Address:        cfg.Address.URL(),
		PollInterval:   cfg.PollInterval.Duration,
		ReportInterval: cfg.ReportInterval.Duration,
		Key:            cfg.Key,
		CryptoKey:      cfg.CryptoKey,
		RateLimit:      cfg.RateLimit,
		LogLevel:       cfg.LogLevel,
	}, nil
}

// === Server Configuration ===

// ServerConfig - the final configuration structure for the server.
type ServerConfig struct {
	Address          string
	LogLevel         string
	StoreInterval    time.Duration
	FileStoragePath  string
	Restore          bool
	DatabaseDSN      string
	EnableRetry      bool   // Enable retry logic for PostgreSQL operations
	Key              string // Key for hash calculation
	CryptoKey        string // Path to PEM file with RSA private key (optional)
	AuditFilePath    string
	AuditURL         string
	AuditHTTPTimeout time.Duration
}

func (c *ServerConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", c.Address)
	enc.AddString("level", c.LogLevel)
	enc.AddDuration("store_interval", c.StoreInterval)
	enc.AddString("storage_path", c.FileStoragePath)
	enc.AddBool("restore", c.Restore)
	enc.AddString("audit_file_path", c.AuditFilePath)
	enc.AddString("audit_url", c.AuditURL)
	enc.AddDuration("audit_http_timeout", c.AuditHTTPTimeout)
	return nil
}

type serverInternalConfig struct {
	Address          Address  `env:"ADDRESS" json:"address"`
	LogLevel         string   `env:"LOG_LEVEL" json:"log_level"`
	StoreInterval    Duration `env:"STORE_INTERVAL" json:"-"`
	FileStoragePath  string   `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore          bool     `env:"RESTORE" json:"-"`
	DatabaseDSN      string   `env:"DATABASE_DSN" json:"database_dsn"`
	EnableRetry      bool     `env:"ENABLE_RETRY" json:"-"`
	Key              string   `env:"KEY" json:"key"`
	CryptoKey        string   `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigPath       string   `env:"CONFIG"`
	AuditFilePath    string   `env:"AUDIT_FILE" json:"audit_file"`
	AuditURL         string   `env:"AUDIT_URL" json:"audit_url"`
	AuditHTTPTimeout Duration `env:"AUDIT_HTTP_TIMEOUT" json:"-"`
}

type serverJSONOverrides struct {
	StoreInterval    *string `json:"store_interval"`
	Restore          *bool   `json:"restore"`
	EnableRetry      *bool   `json:"enable_retry"`
	AuditHTTPTimeout *string `json:"audit_http_timeout"`
}

func resolveConfigPath(args []string) (string, error) {
	var path string
	for i := 0; i < len(args); i++ {
		if args[i] == "-c" || args[i] == "-config" {
			if i+1 >= len(args) {
				return "", fmt.Errorf("missing value for %s", args[i])
			}
			path = args[i+1]
			break
		}
	}
	if envPath := os.Getenv("CONFIG"); envPath != "" {
		return envPath, nil
	}
	return path, nil
}

func loadJSONConfig(path string, dst any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return err
	}
	return nil
}

func applyOptionalBool(src *bool, dst *bool) {
	if src != nil {
		*dst = *src
	}
}

func applyOptionalDuration(src *string, dst *Duration) error {
	if src == nil {
		return nil
	}
	if err := dst.Set(*src); err != nil {
		return err
	}
	return nil
}

func initAgentDefaults(cfg *agentInternalConfig) {
	cfg.Address = Address{Schema: "http", Host: "127.0.0.1", Port: 8080}
	cfg.PollInterval = Duration{Duration: 2 * time.Second}
	cfg.ReportInterval = Duration{Duration: 10 * time.Second}
	cfg.Key = ""
	cfg.RateLimit = 0
	cfg.LogLevel = "info"
}

func initServerDefaults(cfg *serverInternalConfig) {
	cfg.Address = Address{Host: "127.0.0.1", Port: 8080}
	cfg.LogLevel = "info"
	cfg.StoreInterval = Duration{Duration: 300 * time.Second}
	cfg.Restore = true
	cfg.EnableRetry = true
	cfg.Key = ""
	cfg.AuditHTTPTimeout = Duration{Duration: 2 * time.Second}
	cfg.FileStoragePath = "metrics.json"
}

func applyServerJSONOverrides(cfg *serverInternalConfig, overrides serverJSONOverrides) error {
	if err := applyOptionalDuration(overrides.StoreInterval, &cfg.StoreInterval); err != nil {
		return err
	}
	applyOptionalBool(overrides.Restore, &cfg.Restore)
	applyOptionalBool(overrides.EnableRetry, &cfg.EnableRetry)
	if err := applyOptionalDuration(overrides.AuditHTTPTimeout, &cfg.AuditHTTPTimeout); err != nil {
		return err
	}
	return nil
}

// LoadServerConfig loads server configuration from flags and environment (env overrides flags).
func LoadServerConfig() (ServerConfig, error) {
	cfg := serverInternalConfig{}
	initServerDefaults(&cfg)

	configPath, err := resolveConfigPath(os.Args[1:])
	if err != nil {
		return ServerConfig{}, fmt.Errorf("config path parsing error: %w", err)
	}
	cfg.ConfigPath = configPath
	if cfg.ConfigPath != "" {
		raw, err := os.ReadFile(cfg.ConfigPath)
		if err != nil {
			return ServerConfig{}, fmt.Errorf("config file error: %w", err)
		}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return ServerConfig{}, fmt.Errorf("config file error: %w", err)
		}
		var overrides serverJSONOverrides
		if err := json.Unmarshal(raw, &overrides); err != nil {
			return ServerConfig{}, fmt.Errorf("config file error: %w", err)
		}
		if err := applyServerJSONOverrides(&cfg, overrides); err != nil {
			return ServerConfig{}, err
		}
	}
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.Var(&cfg.Address, "a", "server address")
	fs.StringVar(&cfg.ConfigPath, "c", cfg.ConfigPath, "path to JSON config file")
	fs.StringVar(&cfg.ConfigPath, "config", cfg.ConfigPath, "path to JSON config file")
	fs.StringVar(&cfg.LogLevel, "l", cfg.LogLevel, "logging level")
	fs.Var(&cfg.StoreInterval, "i", "server data save interval (seconds or duration)")
	fs.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file path")
	fs.BoolVar(&cfg.Restore, "r", cfg.Restore, "load data from file at startup")
	fs.BoolVar(
		&cfg.EnableRetry,
		"retry",
		cfg.EnableRetry,
		"enable retry logic for PostgreSQL operations (3 retries with intervals 1s, 3s, 5s)",
	)
	fs.StringVar(
		&cfg.DatabaseDSN, "d", cfg.DatabaseDSN,
		"database dsn, example: postgres://user:password@localhost:5432/db?sslmode=disable",
	)
	fs.StringVar(&cfg.Key, "k", cfg.Key, "key for hash calculation")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path to RSA private key PEM for decrypting agent payloads")
	fs.StringVar(&cfg.AuditFilePath, "audit-file", cfg.AuditFilePath, "the path to the file where the audit logs are saved")
	fs.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "the full URL where the audit logs are sent")
	fs.Var(&cfg.AuditHTTPTimeout, "audit-http-timeout", "audit HTTP timeout (seconds or duration)")

	// Flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return ServerConfig{}, fmt.Errorf("flag parsing error: %w", err)
	}

	// Env
	if err := env.Parse(&cfg); err != nil {
		return ServerConfig{}, fmt.Errorf("ENV parsing error: %w", err)
	}

	return ServerConfig{
		Address:          cfg.Address.String(),
		LogLevel:         cfg.LogLevel,
		StoreInterval:    cfg.StoreInterval.Duration,
		FileStoragePath:  cfg.FileStoragePath,
		Restore:          cfg.Restore,
		DatabaseDSN:      cfg.DatabaseDSN,
		EnableRetry:      cfg.EnableRetry,
		Key:              cfg.Key,
		CryptoKey:        cfg.CryptoKey,
		AuditFilePath:    cfg.AuditFilePath,
		AuditURL:         cfg.AuditURL,
		AuditHTTPTimeout: cfg.AuditHTTPTimeout.Duration,
	}, nil
}
