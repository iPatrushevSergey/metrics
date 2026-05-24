// Package trustedsubnet restricts HTTP access to clients in a configured CIDR.
package trustedsubnet

import (
	"net"
	"net/http"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/netutil"
	"github.com/gin-gonic/gin"
)

// TrustedSubnet rejects requests when subnet is set and X-Real-IP is outside it.
func TrustedSubnet(subnet *net.IPNet) gin.HandlerFunc {
	if subnet == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if !netutil.IPInSubnet(subnet, c.GetHeader("X-Real-IP")) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
