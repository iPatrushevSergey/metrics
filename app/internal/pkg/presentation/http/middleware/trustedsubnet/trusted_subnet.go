// Package trustedsubnet restricts HTTP access to clients in a configured CIDR.
package trustedsubnet

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const realIPHeader = "X-Real-IP"

// TrustedSubnet rejects requests when subnet is set and X-Real-IP is outside it.
func TrustedSubnet(subnet *net.IPNet) gin.HandlerFunc {
	if subnet == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		ip := net.ParseIP(strings.TrimSpace(c.GetHeader(realIPHeader)))
		if ip == nil || !subnet.Contains(ip) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
