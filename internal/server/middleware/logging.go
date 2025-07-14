// Package middleware provides HTTP middleware for the server, including request logging.
package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/a2sh3r/sysmetrics/internal/logger"
)

// loggingResponseWriter wraps http.ResponseWriter to capture response status and size.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseStatus int
	responseSize   int
}

// Write writes the response and tracks the size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseSize += size
	return size, err
}

// WriteHeader writes the response status code.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseStatus = statusCode
}

type logEntry struct {
	method   string
	path     string
	status   int
	size     int
	duration time.Duration
}

var logChan = make(chan logEntry, 1000)

func init() {
	go func() {
		for entry := range logChan {
			logger.Log.Info("HTTP request",
				zap.String("method", entry.method),
				zap.String("path", entry.path),
				zap.Int("status", entry.status),
				zap.Int("size", entry.size),
				zap.Duration("duration", entry.duration),
			)
		}
	}()
}

// NewLoggingMiddleware returns a middleware that logs HTTP requests.
func NewLoggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lw := &loggingResponseWriter{
				ResponseWriter: w,
				responseStatus: http.StatusOK,
			}

			next.ServeHTTP(lw, r)

			duration := time.Since(start)

			select {
			case logChan <- logEntry{
				method:   r.Method,
				path:     r.URL.Path,
				status:   lw.responseStatus,
				size:     lw.responseSize,
				duration: duration,
			}:
			default:
			}
		})
	}
}
