// Package config loads server settings from env, flags, and YAML or JSON files.
package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	mapstructure "github.com/go-viper/mapstructure/v2"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/postgres"
)

type Config struct {
	Logger logger.Config   `mapstructure:"logger"`
	Server Server          `mapstructure:"server"`
	GRPC   GRPC            `mapstructure:"grpc"`
	Audit  Audit           `mapstructure:"audit"`
	DB     postgres.Config `mapstructure:"database"`
}

// GRPC holds gRPC server listen settings.
type GRPC struct {
	Address string `mapstructure:"address"`
}

// Audit holds audit fan-out and delivery settings.
type Audit struct {
	AuditSubSize int `mapstructure:"audit_sub_size"`
}

// Server holds server settings.
type Server struct {
	Address          string        `mapstructure:"address"`
	StoreInterval    time.Duration `mapstructure:"store_interval"`
	ShutdownTimeout  time.Duration `mapstructure:"shutdown_timeout"`
	FileStoragePath  string        `mapstructure:"store_file"`
	Restore          bool          `mapstructure:"restore"`
	EnableRetry      bool          `mapstructure:"enable_retry"`
	Key              string        `mapstructure:"key"`
	CryptoKey        string        `mapstructure:"crypto_key"`
	AuditFilePath    string        `mapstructure:"audit_file"`
	AuditURL         string        `mapstructure:"audit_url"`
	AuditHTTPTimeout time.Duration `mapstructure:"audit_http_timeout"`
	TrustedSubnet    string        `mapstructure:"trusted_subnet"`
}

// LoadConfig loads server settings. Priority: env > flags > file > defaults.
func LoadConfig() (Config, error) {
	fs := pflag.NewFlagSet("server", pflag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringP("address", "a", "", "listen address (host:port or http://host:port)")
	fs.StringP("config", "c", "app/configs/server.yaml", "path to YAML or JSON config")
	fs.StringP("trusted-subnet", "t", "", "trusted CIDR for agent X-Real-IP (empty = no check)")
	fs.StringP("log", "l", "", "logging level")
	fs.StringP("store-interval", "i", "", "server data save interval (seconds or duration)")
	fs.StringP("store-file", "f", "", "file path for metric storage")
	fs.BoolP("restore", "r", false, "load data from file at startup")
	fs.Bool("retry", false, "enable retry logic for PostgreSQL operations")
	fs.StringP(
		"database-dsn",
		"d",
		"",
		"database dsn, example: postgres://user:password@localhost:5432/db?sslmode=disable",
	)
	fs.StringP("key", "k", "", "key for hash calculation")
	fs.String("crypto-key", "", "path to RSA private key PEM for decrypting agent payloads")
	fs.String("audit-file", "", "the path to the file where the audit logs are saved")
	fs.String("audit-url", "", "the full URL where the audit logs are sent")
	fs.String("audit-http-timeout", "", "audit HTTP timeout (seconds or duration)")
	fs.Int("audit-sub-size", 0, "audit subscriber channel buffer size per sink")
	fs.String("shutdown-timeout", "", "graceful shutdown timeout (seconds or duration)")
	fs.String("grpc-address", "", "gRPC listen address (host:port)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return Config{}, fmt.Errorf("flag parsing error: %w", err)
	}

	dotenvPath, dotenvLoaded, err := loadDotEnv()
	if err != nil {
		return Config{}, fmt.Errorf("load .env error: %w", err)
	}
	if dotenvLoaded {
		_, _ = fmt.Fprintf(os.Stderr, "config: loaded dotenv file %s\n", dotenvPath)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "config: dotenv file not found, continuing with env/yaml/defaults")
	}

	v := viper.New()
	setDefaults(v)

	configPath, _ := fs.GetString("config")
	if _, ok := os.LookupEnv("CONFIG"); ok {
		configPath = strings.TrimSpace(os.Getenv("CONFIG"))
	}
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_, configEnvSet := os.LookupEnv("CONFIG")
			configPathFromEnv := configEnvSet && strings.TrimSpace(os.Getenv("CONFIG")) != ""
			if fs.Changed("config") || configPathFromEnv {
				return Config{}, fmt.Errorf("config file not found: %s", configPath)
			}
			_, _ = fmt.Fprintf(os.Stderr, "config: default config file %s not found, continuing with env/defaults\n", configPath)
		} else {
			return Config{}, fmt.Errorf("read config file: %w", err)
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "config: loaded config file %s\n", v.ConfigFileUsed())
	}

	bindEnv(v)
	if err := applyFlagsWhenEnvUnset(v, fs); err != nil {
		return Config{}, fmt.Errorf("apply flags: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(durationDecodeHook())); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := finalizeConfig(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// durationDecodeHook maps incoming scalars to time.
func durationDecodeHook() mapstructure.DecodeHookFunc {
	durType := reflect.TypeOf(time.Duration(0))
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if to != durType {
			return data, nil
		}
		d, err := parseDuration(data)
		if err != nil {
			return nil, err
		}
		return d, nil
	}
}

// parseListenAddress parses the listen address.
func parseListenAddress(raw string) (string, error) {
	var addr Address
	if err := addr.Set(raw); err != nil {
		return "", err
	}
	return addr.String(), nil
}

// parseDuration parses the duration.
func parseDuration(raw any) (time.Duration, error) {
	var d Duration
	switch v := raw.(type) {
	case string:
		if err := d.Set(v); err != nil {
			return 0, err
		}
		return d.Duration, nil
	case int:
		return time.Duration(v) * time.Second, nil
	case int32:
		return time.Duration(v) * time.Second, nil
	case int64:
		return time.Duration(v) * time.Second, nil
	case float64:
		return time.Duration(v) * time.Second, nil
	default:
		return 0, fmt.Errorf("unsupported duration type %T", raw)
	}
}

// loadDotEnv loads the .env file if it exists.
func loadDotEnv() (string, bool, error) {
	for _, file := range []string{"app/.env", ".env"} {
		if _, err := os.Stat(file); err == nil {
			if err := godotenv.Load(file); err != nil {
				return "", false, fmt.Errorf("failed to load %s: %w", file, err)
			}
			return file, true, nil
		}
	}
	return "", false, nil
}

// setDefaults sets the default values for the configuration.
func setDefaults(v *viper.Viper) {
	v.SetDefault("logger.level", "info")

	v.SetDefault("server.address", "127.0.0.1:8080")
	v.SetDefault("server.store_interval", "300s")
	v.SetDefault("server.shutdown_timeout", "10s")
	v.SetDefault("server.store_file", "metrics.json")
	v.SetDefault("server.restore", true)
	v.SetDefault("server.enable_retry", true)
	v.SetDefault("server.key", "")
	v.SetDefault("server.crypto_key", "")
	v.SetDefault("server.audit_file", "")
	v.SetDefault("server.audit_url", "")
	v.SetDefault("server.audit_http_timeout", "2s")
	v.SetDefault("server.trusted_subnet", "")
	v.SetDefault("grpc.address", "127.0.0.1:3000")

	v.SetDefault("audit.audit_sub_size", 500)

	v.SetDefault("database.uri", "")
	v.SetDefault("database.max_conns", 25)
	v.SetDefault("database.min_conns", 5)
	v.SetDefault("database.max_conn_life", "1h")
	v.SetDefault("database.max_conn_idle", "30m")
	v.SetDefault("database.health_check", "1m")
	v.SetDefault("database.retry.max_retries", 3)
	v.SetDefault("database.retry.base_delay", "100ms")
	v.SetDefault("database.retry.max_delay", "2s")
}

// bindEnv binds the environment variables to the configuration.
func bindEnv(v *viper.Viper) {
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	_ = v.BindEnv("server.address", "ADDRESS")
	_ = v.BindEnv("logger.level", "LOG_LEVEL")
	_ = v.BindEnv("server.store_interval", "STORE_INTERVAL")
	_ = v.BindEnv("server.shutdown_timeout", "SHUTDOWN_TIMEOUT")
	_ = v.BindEnv("server.store_file", "FILE_STORAGE_PATH")
	_ = v.BindEnv("server.restore", "RESTORE")
	_ = v.BindEnv("server.enable_retry", "ENABLE_RETRY")
	_ = v.BindEnv("server.key", "KEY")
	_ = v.BindEnv("server.crypto_key", "CRYPTO_KEY")
	_ = v.BindEnv("server.audit_file", "AUDIT_FILE")
	_ = v.BindEnv("server.audit_url", "AUDIT_URL")
	_ = v.BindEnv("server.audit_http_timeout", "AUDIT_HTTP_TIMEOUT")
	_ = v.BindEnv("server.trusted_subnet", "TRUSTED_SUBNET")
	_ = v.BindEnv("grpc.address", "GRPC_ADDRESS")
	_ = v.BindEnv("audit.audit_sub_size", "AUDIT_SUB_SIZE")

	_ = v.BindEnv("database.uri", "DATABASE_DSN")
	_ = v.BindEnv("database.max_conns", "DB_MAX_CONNS")
	_ = v.BindEnv("database.min_conns", "DB_MIN_CONNS")
	_ = v.BindEnv("database.max_conn_life", "DB_MAX_CONN_LIFE")
	_ = v.BindEnv("database.max_conn_idle", "DB_MAX_CONN_IDLE")
	_ = v.BindEnv("database.health_check", "DB_HEALTH_CHECK")
	_ = v.BindEnv("database.retry.max_retries", "DB_RETRY_MAX_RETRIES")
	_ = v.BindEnv("database.retry.base_delay", "DB_RETRY_BASE_DELAY")
	_ = v.BindEnv("database.retry.max_delay", "DB_RETRY_MAX_DELAY")
}

// applyFlagsWhenEnvUnset sets viper from CLI flags only when the matching env var is absent (env beats flags).
func applyFlagsWhenEnvUnset(v *viper.Viper, fs *pflag.FlagSet) error {
	for _, row := range []struct{ key, env, flag string }{
		{"server.address", "ADDRESS", "address"},
		{"logger.level", "LOG_LEVEL", "log"},
		{"server.store_interval", "STORE_INTERVAL", "store-interval"},
		{"server.store_file", "FILE_STORAGE_PATH", "store-file"},
		{"server.key", "KEY", "key"},
		{"server.crypto_key", "CRYPTO_KEY", "crypto-key"},
		{"server.audit_file", "AUDIT_FILE", "audit-file"},
		{"server.audit_url", "AUDIT_URL", "audit-url"},
		{"server.audit_http_timeout", "AUDIT_HTTP_TIMEOUT", "audit-http-timeout"},
		{"server.shutdown_timeout", "SHUTDOWN_TIMEOUT", "shutdown-timeout"},
		{"server.trusted_subnet", "TRUSTED_SUBNET", "trusted-subnet"},
		{"grpc.address", "GRPC_ADDRESS", "grpc-address"},
	} {
		if _, ok := os.LookupEnv(row.env); ok {
			continue
		}
		if !fs.Changed(row.flag) {
			continue
		}
		val, err := fs.GetString(row.flag)
		if err != nil {
			return fmt.Errorf("flag %s: %w", row.flag, err)
		}
		v.Set(row.key, val)
	}

	if _, ok := os.LookupEnv("DATABASE_DSN"); !ok && fs.Changed("database-dsn") {
		dsnVal, err := fs.GetString("database-dsn")
		if err != nil {
			return fmt.Errorf("flag database-dsn: %w", err)
		}
		v.Set("database.uri", dsnVal)
	}

	if _, ok := os.LookupEnv("RESTORE"); !ok && fs.Changed("restore") {
		restoreVal, err := fs.GetBool("restore")
		if err != nil {
			return fmt.Errorf("flag restore: %w", err)
		}
		v.Set("server.restore", restoreVal)
	}
	if _, ok := os.LookupEnv("ENABLE_RETRY"); !ok && fs.Changed("retry") {
		retryVal, err := fs.GetBool("retry")
		if err != nil {
			return fmt.Errorf("flag retry: %w", err)
		}
		v.Set("server.enable_retry", retryVal)
	}
	if _, ok := os.LookupEnv("AUDIT_SUB_SIZE"); !ok && fs.Changed("audit-sub-size") {
		subSize, err := fs.GetInt("audit-sub-size")
		if err != nil {
			return fmt.Errorf("flag audit-sub-size: %w", err)
		}
		v.Set("audit.audit_sub_size", subSize)
	}
	return nil
}

// finalizeConfig validates and normalizes the listen address and string fields.
func finalizeConfig(cfg *Config) error {
	s := &cfg.Server
	addr, err := parseListenAddress(s.Address)
	if err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}
	s.Address = addr

	s.Key = strings.TrimSpace(s.Key)
	s.CryptoKey = strings.TrimSpace(s.CryptoKey)
	s.FileStoragePath = strings.TrimSpace(s.FileStoragePath)
	s.AuditFilePath = strings.TrimSpace(s.AuditFilePath)
	s.AuditURL = strings.TrimSpace(s.AuditURL)
	s.TrustedSubnet = strings.TrimSpace(s.TrustedSubnet)
	if s.TrustedSubnet != "" {
		if _, _, err := net.ParseCIDR(s.TrustedSubnet); err != nil {
			return fmt.Errorf("invalid trusted_subnet: %w", err)
		}
	}

	if addr := strings.TrimSpace(cfg.GRPC.Address); addr != "" {
		parsed, err := parseListenAddress(addr)
		if err != nil {
			return fmt.Errorf("invalid grpc address: %w", err)
		}
		cfg.GRPC.Address = parsed
	}

	cfg.DB.Pool.URI = strings.TrimSpace(cfg.DB.Pool.URI)

	if s.StoreInterval < 0 {
		return fmt.Errorf("store interval must be >= 0, got %s", s.StoreInterval)
	}
	if s.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be > 0, got %s", s.ShutdownTimeout)
	}
	if s.AuditHTTPTimeout < 0 {
		return fmt.Errorf("audit http timeout must be >= 0, got %s", s.AuditHTTPTimeout)
	}
	if cfg.Audit.AuditSubSize <= 0 {
		return fmt.Errorf("audit sub size must be > 0, got %d", cfg.Audit.AuditSubSize)
	}
	return nil
}
