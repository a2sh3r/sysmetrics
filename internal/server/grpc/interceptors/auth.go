// Package interceptors provides gRPC interceptors for authentication, logging, and other middleware functionality.
package interceptors

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/a2sh3r/sysmetrics/internal/hash"
)

// AuthInterceptor provides authentication for gRPC requests
func AuthInterceptor(secretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if secretKey == "" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		hashValues := md.Get("hashsha256")
		if len(hashValues) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing hash")
		}

		requestData, err := json.Marshal(req)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to serialize request")
		}

		if err := hash.VerifyHash(string(requestData), secretKey, hashValues[0]); err != nil {
			return nil, status.Error(codes.Unauthenticated, "hash verification failed")
		}

		return handler(ctx, req)
	}
}
