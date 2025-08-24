// Package interceptors provides gRPC interceptors for authentication, logging, and other middleware functionality.
package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/a2sh3r/sysmetrics/internal/logger"
)

// LoggingInterceptor provides logging for gRPC requests
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		peer := getPeerFromContext(ctx)

		logger.Log.Info("gRPC request started",
			zap.String("method", info.FullMethod),
			zap.String("peer", peer))

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			}
		}

		logger.Log.Info("gRPC request completed",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("status", statusCode.String()),
			zap.Error(err))

		return resp, err
	}
}

func getPeerFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "unknown"
	}

	if peer := md.Get("x-real-ip"); len(peer) > 0 {
		return peer[0]
	}

	return "unknown"
}
