package interceptors

import (
	"context"
	"net"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/netutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestTrustedSubnet_allowed(t *testing.T) {
	_, subnet, err := net.ParseCIDR("192.168.0.0/16")
	require.NoError(t, err)

	ic := TrustedSubnet(subnet)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(netutil.RealIPHeader, "192.168.1.5"))

	called := false
	_, err = ic(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		called = true
		return "ok", nil
	})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestTrustedSubnet_forbidden(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	ic := TrustedSubnet(subnet)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(netutil.RealIPHeader, "203.0.113.1"))

	_, err = ic(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		t.Fatal("handler must not run")
		return nil, nil
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestTrustedSubnet_disabled(t *testing.T) {
	ic := TrustedSubnet(nil)
	_, err := ic(context.Background(), nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		return nil, nil
	})
	require.NoError(t, err)
}
