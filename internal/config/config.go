package config

import (
	"errors"
	"flag"
	"fmt"
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
	Host string
	Port int
}

// String implements the interface flag.Value
func (a *Address) String() string {
	if a.Host == "" && a.Port == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// Set implements the interface flag.Value
func (a *Address) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form 'host:port'")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

func (a *Address) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

// === Agent Configuration ===

// AgentConfig - the final configuration structure for the agent.
type AgentConfig struct {
	Address        string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type agentInternalConfig struct {
	Address        Address `env:"ADDRESS"`
	PollInterval   int     `env:"POLL_INTERVAL"`
	ReportInterval int     `env:"REPORT_INTERVAL"`
}

// Environment variables take precedence over flags.
func LoadAgentConfig() (AgentConfig, error) {
	cfg := agentInternalConfig{}

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	// Default
	cfg.Address = Address{Host: "127.0.0.1", Port: 8080}
	fs.Var(&cfg.Address, "a", "server address (default '127.0.0.1:8080')")
	fs.IntVar(&cfg.ReportInterval, "r", 10, "frequency of sending metrics (default 10s)")
	fs.IntVar(&cfg.PollInterval, "p", 2, "frequency of metrics polling (default 2s)")

	// Flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return AgentConfig{}, fmt.Errorf("flag parsing error: %w", err)
	}

	// Env
	if err := env.Parse(&cfg); err != nil {
		return AgentConfig{}, fmt.Errorf("ENV parsing error: %w", err)
	}

	finalCfg := AgentConfig{
		Address:        "http://" + cfg.Address.String(),
		PollInterval:   time.Duration(cfg.PollInterval) * time.Second,
		ReportInterval: time.Duration(cfg.ReportInterval) * time.Second,
	}

	return finalCfg, nil
}

// === Server Configuration ===

// ServerConfig - the final configuration structure for the server.
type ServerConfig struct {
	Address  string
	LogLevel string
}

func (c *ServerConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", c.Address)
	enc.AddString("level", c.LogLevel)
	return nil
}

type serverInternalConfig struct {
	Address  Address `env:"ADDRESS"`
	LogLevel string  `env:"LOG_LEVEL"`
}

// Environment variables take precedence over flags.
func LoadServerConfig() (ServerConfig, error) {
	cfg := serverInternalConfig{}

	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	// Default
	cfg.Address = Address{Host: "127.0.0.1", Port: 8080}
	fs.Var(&cfg.Address, "a", "server address (default '127.0.0.1:8080')")
	fs.StringVar(&cfg.LogLevel, "l", "info", "logging level (default 'info')")

	// Flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return ServerConfig{}, fmt.Errorf("flag parsing error: %w", err)
	}

	// Env
	if err := env.Parse(&cfg); err != nil {
		return ServerConfig{}, fmt.Errorf("flag parsing error: %w", err)
	}

	finalCfg := ServerConfig{
		Address:  cfg.Address.String(),
		LogLevel: cfg.LogLevel,
	}

	return finalCfg, nil
}
