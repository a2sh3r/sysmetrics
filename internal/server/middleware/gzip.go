package middleware

import (
	"compress/gzip"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func NewGzipMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") == "gzip" {
				gr, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Failed to read gzip body", http.StatusBadRequest)
					logger.Log.Error("Failed to read gzip body", zap.Error(err))
					return
				}
				defer func() {
					if err := gr.Close(); err != nil {
						log.Printf("failed to close gz.Close: %v", err)
					}
				}()

				r.Body = io.NopCloser(gr)
			}

			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Encoding", "gzip")
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, "Failed to create gzip writer", http.StatusInternalServerError)
				logger.Log.Error("Failed to create gzip writer", zap.Error(err))
				return
			}
			defer func() {
				if err := gz.Close(); err != nil {
					log.Printf("failed to close gz.Close: %v", err)
				}
			}()

			grw := &gzipWriter{
				ResponseWriter: w,
				writer:         gz,
			}

			next.ServeHTTP(grw, r)
		})
	}
}
