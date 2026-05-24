package netutil

import (
	"net"
	"strings"
)

// IPInSubnet checks if ip is inside subnet; true when subnet is nil.
func IPInSubnet(subnet *net.IPNet, ip string) bool {
	if subnet == nil {
		return true
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	return parsed != nil && subnet.Contains(parsed)
}
