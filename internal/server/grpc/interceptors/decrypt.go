// Package interceptors provides gRPC interceptors for authentication, logging, and other middleware functionality.
package interceptors

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/a2sh3r/sysmetrics/internal/crypto"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"go.uber.org/zap"
)

// DecryptInterceptor provides decryption for gRPC requests
func DecryptInterceptor(decryptor *crypto.Decryptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if decryptor == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		encryptedValues := md.Get("x-encrypted")
		if len(encryptedValues) == 0 || encryptedValues[0] != "true" {
			return handler(ctx, req)
		}

		serializedReq, err := json.Marshal(req)
		if err != nil {
			logger.Log.Error("Failed to serialize request for decryption", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to process encrypted request")
		}

		decryptedData, err := decryptor.Decrypt(serializedReq)
		if err != nil {
			logger.Log.Error("Failed to decrypt request", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "Failed to decrypt request")
		}

		decryptedReq, err := deserializeRequest(decryptedData, req)
		if err != nil {
			logger.Log.Error("Failed to deserialize decrypted request", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "Failed to process decrypted request")
		}

		logger.Log.Debug("Successfully decrypted gRPC request")

		return handler(ctx, decryptedReq)
	}
}

func deserializeRequest(data []byte, originalReq interface{}) (interface{}, error) {
	return nil, nil
}
