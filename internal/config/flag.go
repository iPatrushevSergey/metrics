package config

import (
	"errors"
	"flag"
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

func AgentParseFlags() AgentOptions {
	options := AgentOptions{
		NetAddress: NetAddress{Host: "localhost", Port: 8080},
	}

	flag.Var(&options.NetAddress, "a", "address and port to run server (default '127.0.0.1:8080')")
	flag.IntVar(&options.ReportInterval, "r", 10, "frequency of sending metrics (default 10s)")
	flag.IntVar(&options.PollInterval, "p", 2, "frequency of metrics polling (default 2s)")

	flag.Parse()
	return options
}

func ServerParseFlags() ServerOptions {
	options := ServerOptions{
		NetAddress: NetAddress{Host: "localhost", Port: 8080},
	}

	flag.Var(&options.NetAddress, "a", "address and port to run server (default '127.0.0.1:8080')")

	flag.Parse()
	return options
}
