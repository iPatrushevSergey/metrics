package netutil

import (
	"net"
	"strings"
)

// RealIPHeader is the agent client IP in HTTP headers and gRPC metadata.
const RealIPHeader = "X-Real-IP"

// IPInSubnet checks if ip is inside subnet; true when subnet is nil.
func IPInSubnet(subnet *net.IPNet, ip string) bool {
	if subnet == nil {
		return true
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	return parsed != nil && subnet.Contains(parsed)
}
