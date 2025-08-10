package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/a2sh3r/sysmetrics/internal/crypto"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"go.uber.org/zap"
)

// DecryptMiddleware is a middleware that decrypts encrypted request bodies.
type DecryptMiddleware struct {
	decryptor *crypto.Decryptor
}

// NewDecryptMiddleware creates a new DecryptMiddleware.
func NewDecryptMiddleware(privateKeyPath string) (*DecryptMiddleware, error) {
	if privateKeyPath == "" {
		return &DecryptMiddleware{}, nil
	}

	decryptor, err := crypto.NewDecryptor(privateKeyPath)
	if err != nil {
		return nil, err
	}

	return &DecryptMiddleware{decryptor: decryptor}, nil
}

// DecryptBody is a middleware function that decrypts the request body if it's encrypted.
func (dm *DecryptMiddleware) DecryptBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Encrypted") == "true" {
			if dm.decryptor == nil {
				logger.Log.Error("Received encrypted request but no decryptor configured")
				http.Error(w, "Server not configured for decryption", http.StatusInternalServerError)
				return
			}

			encryptedBody, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("Failed to read encrypted body", zap.Error(err))
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			if closeErr := r.Body.Close(); closeErr != nil {
				logger.Log.Warn("Failed to close request body", zap.Error(closeErr))
			}

			decryptedBody, err := dm.decryptor.Decrypt(encryptedBody)
			if err != nil {
				logger.Log.Error("Failed to decrypt request body", zap.Error(err))
				http.Error(w, "Failed to decrypt request body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decryptedBody))
			r.ContentLength = int64(len(decryptedBody))

			logger.Log.Debug("Successfully decrypted request body", zap.Int("size", len(decryptedBody)))
		}

		next.ServeHTTP(w, r)
	})
}
