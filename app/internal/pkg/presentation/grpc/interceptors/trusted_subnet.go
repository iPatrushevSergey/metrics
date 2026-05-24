package interceptors

import (
	"context"
	"net"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TrustedSubnet rejects unary RPCs when subnet is set and X-Real-IP is outside it.
func TrustedSubnet(subnet *net.IPNet) grpc.UnaryServerInterceptor {
	if subnet == nil {
		return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		realIP := ""
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get(netutil.RealIPHeader); len(vals) > 0 {
				realIP = vals[0]
			}
		}
		if !netutil.IPInSubnet(subnet, realIP) {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}
		return handler(ctx, req)
	}
}
