package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	mapstructure "github.com/go-viper/mapstructure/v2"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/logger"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/adapters/metrics_client"
)

type Config struct {
	Logger logger.Config `mapstructure:"logger"`
	Agent  Agent         `mapstructure:"agent"`
}

// Agent holds collector settings.
type Agent struct {
	metrics_client.MetricsClientConfig `mapstructure:",squash"`
	PollInterval                       time.Duration `mapstructure:"poll_interval"`
	ReportInterval                     time.Duration `mapstructure:"report_interval"`
	// RateLimit is the worker-pool size: max simultaneous outbound metric batch RPCs toward the server. Not "requests per second".
	RateLimit int    `mapstructure:"rate_limit"`
	Key       string `mapstructure:"key"`
	CryptoKey string `mapstructure:"crypto_key"`
}

// LoadConfig loads agent settings. Priority: flags > env > file > defaults.
func LoadConfig() (Config, error) {
	fs := pflag.NewFlagSet("agent", pflag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringP("address", "a", "", "metrics server URL or host:port")
	fs.StringP("config", "c", "app/configs/agent.yaml", "path to YAML config")
	fs.StringP("poll-interval", "p", "", "poll interval (seconds or duration, e.g. 2s)")
	fs.StringP("report-interval", "r", "", "report interval (seconds or duration)")
	fs.StringP("key", "k", "", "key for hash calculation")
	fs.String("http-timeout", "", "metrics server HTTP client timeout (duration)")
	fs.String("crypto-key", "", "path to RSA public key PEM for payload encryption")
	fs.IntP("rate-limit", "l", 0, "concurrent batch workers (0 = sequential)")
	fs.String("log", "", "logging level")

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
	configEnvRaw, configEnvOk := os.LookupEnv("CONFIG")
	configEnv := strings.TrimSpace(configEnvRaw)
	if !fs.Changed("config") && configEnvOk && configEnv != "" {
		configPath = configEnv
	}
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if fs.Changed("config") || (configEnvOk && configEnv != "") {
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
	if err := bindFlags(v, fs); err != nil {
		return Config{}, fmt.Errorf("bind flags: %w", err)
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

// parseURLAddress parses the URL address.
func parseURLAddress(raw string) (string, error) {
	var addr Address
	if err := addr.Set(raw); err != nil {
		return "", err
	}
	return addr.URL(), nil
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

	v.SetDefault("agent.address", "http://127.0.0.1:8080")
	v.SetDefault("agent.http_timeout", "2s")
	v.SetDefault("agent.poll_interval", "2s")
	v.SetDefault("agent.report_interval", "10s")
	v.SetDefault("agent.key", "")
	v.SetDefault("agent.crypto_key", "")
	v.SetDefault("agent.rate_limit", 0)
}

// bindEnv binds the environment variables to the configuration.
func bindEnv(v *viper.Viper) {
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	_ = v.BindEnv("agent.address", "ADDRESS")
	_ = v.BindEnv("agent.http_timeout", "HTTP_TIMEOUT")
	_ = v.BindEnv("agent.poll_interval", "POLL_INTERVAL")
	_ = v.BindEnv("agent.report_interval", "REPORT_INTERVAL")
	_ = v.BindEnv("agent.key", "KEY")
	_ = v.BindEnv("agent.crypto_key", "CRYPTO_KEY")
	_ = v.BindEnv("agent.rate_limit", "RATE_LIMIT")
	_ = v.BindEnv("logger.level", "LOG_LEVEL")
}

// bindFlags binds the command line flags to the configuration.
func bindFlags(v *viper.Viper, fs *pflag.FlagSet) error {
	bindings := map[string]string{
		"agent.address":         "address",
		"agent.http_timeout":    "http-timeout",
		"agent.poll_interval":   "poll-interval",
		"agent.report_interval": "report-interval",
		"agent.key":             "key",
		"agent.crypto_key":      "crypto-key",
		"agent.rate_limit":      "rate-limit",
		"logger.level":          "log",
	}
	for key, flagName := range bindings {
		f := fs.Lookup(flagName)
		if f == nil {
			return fmt.Errorf("flag not found: %s", flagName)
		}
		if err := v.BindPFlag(key, f); err != nil {
			return err
		}
	}
	return nil
}

// finalizeConfig validates required fields and normalizes metrics server address.
func finalizeConfig(cfg *Config) error {
	a := &cfg.Agent
	addr, err := parseURLAddress(a.MetricsClientConfig.Address)
	if err != nil {
		return fmt.Errorf("invalid metrics server address: %w", err)
	}
	a.MetricsClientConfig.Address = addr

	a.Key = strings.TrimSpace(a.Key)
	a.CryptoKey = strings.TrimSpace(a.CryptoKey)

	if a.RateLimit < 0 {
		return fmt.Errorf("rate limit must be >= 0, got %d", a.RateLimit)
	}
	return nil
}
