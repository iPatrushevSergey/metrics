package config

import (
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

// === Custom Flag Type ===

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
}

type agentInternalConfig struct {
	Address        Address  `env:"ADDRESS"`
	PollInterval   Duration `env:"POLL_INTERVAL"`
	ReportInterval Duration `env:"REPORT_INTERVAL"`
}

// Environment variables take precedence over flags.
func LoadAgentConfig() (AgentConfig, error) {
	cfg := agentInternalConfig{}

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	// Default
	cfg.Address = Address{Schema: "http", Host: "127.0.0.1", Port: 8080}
	cfg.PollInterval = Duration{Duration: 2 * time.Second}
	cfg.ReportInterval = Duration{Duration: 10 * time.Second}
	fs.Var(&cfg.Address, "a", "server address")
	fs.Var(&cfg.ReportInterval, "r", "frequency of sending metrics (seconds or duration)")
	fs.Var(&cfg.PollInterval, "p", "frequency of metrics polling (seconds or duration)")

	// Flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return AgentConfig{}, fmt.Errorf("flag parsing error: %w", err)
	}

	// Env
	if err := env.Parse(&cfg); err != nil {
		return AgentConfig{}, fmt.Errorf("ENV parsing error: %w", err)
	}

	finalCfg := AgentConfig{
		Address:        cfg.Address.URL(),
		PollInterval:   cfg.PollInterval.Duration,
		ReportInterval: cfg.ReportInterval.Duration,
	}

	return finalCfg, nil
}

// === Server Configuration ===

// ServerConfig - the final configuration structure for the server.
type ServerConfig struct {
	Address         string
	LogLevel        string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
}

func (c *ServerConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", c.Address)
	enc.AddString("level", c.LogLevel)
	enc.AddDuration("store_interval", c.StoreInterval)
	enc.AddString("storage_path", c.FileStoragePath)
	enc.AddBool("restore", c.Restore)
	return nil
}

type serverInternalConfig struct {
	Address         Address  `env:"ADDRESS"`
	LogLevel        string   `env:"LOG_LEVEL"`
	StoreInterval   Duration `env:"STORE_INTERVAL"`
	FileStoragePath string   `env:"FILE_STORAGE_PATH"`
	Restore         bool     `env:"RESTORE"`
	DatabaseDSN     string   `env:"DATABASE_DSN"`
}

// Environment variables take precedence over flags.
func LoadServerConfig() (ServerConfig, error) {
	cfg := serverInternalConfig{}

	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	// Default
	cfg.Address = Address{Host: "127.0.0.1", Port: 8080}
	cfg.StoreInterval = Duration{Duration: 300 * time.Second}
	fs.Var(&cfg.Address, "a", "server address")
	fs.StringVar(&cfg.LogLevel, "l", "info", "logging level")
	fs.Var(&cfg.StoreInterval, "i", "server data save interval (seconds or duration)")
	fs.StringVar(&cfg.FileStoragePath, "f", "metrics.json", "file path")
	fs.BoolVar(&cfg.Restore, "r", true, "load data from file at startup")
	fs.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn")

	// Flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return ServerConfig{}, fmt.Errorf("flag parsing error: %w", err)
	}

	// Env
	if err := env.Parse(&cfg); err != nil {
		return ServerConfig{}, fmt.Errorf("ENV parsing error: %w", err)
	}

	finalCfg := ServerConfig{
		Address:         cfg.Address.String(),
		LogLevel:        cfg.LogLevel,
		StoreInterval:   cfg.StoreInterval.Duration,
		FileStoragePath: cfg.FileStoragePath,
		Restore:         cfg.Restore,
		DatabaseDSN:     cfg.DatabaseDSN,
	}

	return finalCfg, nil
}
