// Package interceptors provides gRPC interceptors for authentication, logging, and other middleware functionality.
package interceptors

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/a2sh3r/sysmetrics/internal/logger"
	"go.uber.org/zap"
)

// IPCheckInterceptor provides IP filtering for gRPC requests
func IPCheckInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trustedSubnet == "" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "unable to get metadata")
		}

		realIPValues := md.Get("x-real-ip")
		if len(realIPValues) == 0 {
			logger.Log.Error("X-Real-IP metadata is missing")
			return nil, status.Error(codes.PermissionDenied, "X-Real-IP metadata is required")
		}

		realIP := realIPValues[0]
		ip := net.ParseIP(realIP)
		if ip == nil {
			logger.Log.Error("Invalid IP address", zap.String("ip", realIP))
			return nil, status.Error(codes.PermissionDenied, "Invalid IP address")
		}

		_, subnet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			logger.Log.Error("Invalid trusted subnet", zap.Error(err))
			return nil, status.Error(codes.Internal, "Server configuration error")
		}

		if !subnet.Contains(ip) {
			logger.Log.Error("IP not in trusted subnet",
				zap.String("ip", realIP),
				zap.String("subnet", trustedSubnet))
			return nil, status.Error(codes.PermissionDenied, "Access denied: IP not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
