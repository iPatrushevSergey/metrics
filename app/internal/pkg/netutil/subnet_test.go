package netutil

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPInSubnet_nilSubnet(t *testing.T) {
	assert.True(t, IPInSubnet(nil, "203.0.113.1"))
}

func TestIPInSubnet_allowed(t *testing.T) {
	_, subnet, err := net.ParseCIDR("192.168.0.0/16")
	require.NoError(t, err)
	assert.True(t, IPInSubnet(subnet, "192.168.1.10"))
}

func TestIPInSubnet_forbidden(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)
	assert.False(t, IPInSubnet(subnet, "203.0.113.1"))
}

func TestIPInSubnet_invalidIP(t *testing.T) {
	_, subnet, err := net.ParseCIDR("127.0.0.0/8")
	require.NoError(t, err)
	assert.False(t, IPInSubnet(subnet, "not-an-ip"))
}
