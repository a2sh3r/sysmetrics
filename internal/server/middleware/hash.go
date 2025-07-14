package middleware

import (
	"bytes"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/hash"
	"github.com/a2sh3r/sysmetrics/internal/logger"
)

const HashHeader = "HashSHA256"

func NewHashMiddleware(cfg *config.ServerConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if cfg.SecretKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("Failed to read request body", zap.Error(err))
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			err = r.Body.Close()
			if err != nil {
				logger.Log.Error("Failed to close request body", zap.Error(err))
				http.Error(w, "Failed to close request body", http.StatusBadRequest)
				return
			}

			gotHash := r.Header.Get(HashHeader)
			if gotHash != "" {
				if err := hash.VerifyHash(string(body), cfg.SecretKey, gotHash); err != nil {
					logger.Log.Error("Hash verification failed", zap.Error(err))
					http.Error(w, "Hash verification failed", http.StatusBadRequest)
					return
				}
			}

			rw := &hashResponseWriter{
				ResponseWriter: w,
				secretKey:      cfg.SecretKey,
			}

			r.Body = io.NopCloser(io.Reader(bytes.NewReader(body)))

			next.ServeHTTP(rw, r)
		})
	}
}

type hashResponseWriter struct {
	http.ResponseWriter
	secretKey     string
	body          []byte
	headerWritten bool
	statusCode    int
}

func (rw *hashResponseWriter) Write(b []byte) (int, error) {
	rw.body = b
	if rw.secretKey != "" {
		calculateHash := hash.CalculateHash(string(b), rw.secretKey)
		rw.Header().Set(HashHeader, calculateHash)
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *hashResponseWriter) WriteHeader(statusCode int) {
	if !rw.headerWritten {
		rw.statusCode = statusCode
		rw.headerWritten = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}
