package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Address contains the schema, host, and port.
type Address struct {
	Schema string
	Host   string
	Port   int
}

// Set parses address string into schema, host and port.
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

// String returns host and port.
func (a *Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// UnmarshalText decodes text using set logic.
func (a *Address) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

// URL returns schema://host:port string.
func (a *Address) URL() string {
	return fmt.Sprintf("%s://%s:%d", a.Schema, a.Host, a.Port)
}

// Duration embeds time.Duration.
type Duration struct {
	time.Duration
}

// Set parses integer or duration string seconds.
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

// String returns duration string.
func (d *Duration) String() string {
	return d.Duration.String()
}

// UnmarshalText decodes text using set logic.
func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}
