package trustedsubnet

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTrustedSubnet_disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TrustedSubnet(nil))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTrustedSubnet_allowed(t *testing.T) {
	_, subnet, err := net.ParseCIDR("192.168.0.0/16")
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TrustedSubnet(subnet))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(realIPHeader, "192.168.1.10")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTrustedSubnet_forbidden(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/8")
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TrustedSubnet(subnet))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(realIPHeader, "203.0.113.1")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTrustedSubnet_missingHeader(t *testing.T) {
	_, subnet, err := net.ParseCIDR("127.0.0.0/8")
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TrustedSubnet(subnet))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
