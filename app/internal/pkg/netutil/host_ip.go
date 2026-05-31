// Package netutil provides helpers for network operations.
package netutil

import (
	"fmt"
	"net"
)

// HostIPv4 returns the first non-loopback IPv4 address of the host.
func HostIPv4() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		ip4 := ipNet.IP.To4()
		if ip4 == nil {
			continue
		}
		return ip4.String(), nil
	}
	return "", fmt.Errorf("no network IPv4 address found")
}
