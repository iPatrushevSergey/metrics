package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

type AgentOptions struct {
	NetAddress     NetAddress
	PollInterval   int
	ReportInterval int
}

type ServerOptions struct {
	NetAddress NetAddress
}

type NetAddress struct {
	Host string
	Port int
}

func (a *NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

func AgentParseFlags() (AgentOptions, error) {
	options := AgentOptions{
		NetAddress: NetAddress{Host: "127.0.0.1", Port: 8080},
	}

	fs := flag.NewFlagSet("agent", flag.ExitOnError)

	fs.Var(&options.NetAddress, "a", "address and port to run server (default '127.0.0.1:8080')")
	fs.IntVar(&options.ReportInterval, "r", 10, "frequency of sending metrics (default 10s)")
	fs.IntVar(&options.PollInterval, "p", 2, "frequency of metrics polling (default 2s)")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return options, err
	}

	return options, nil
}

func ServerParseFlags() (ServerOptions, error) {
	options := ServerOptions{
		NetAddress: NetAddress{Host: "127.0.0.1", Port: 8080},
	}

	fs := flag.NewFlagSet("server", flag.ExitOnError)

	fs.Var(&options.NetAddress, "a", "address and port to run server (default '127.0.0.1:8080')")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return options, err
	}

	return options, nil
}
