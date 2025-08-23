package middleware

import (
	"net"
	"net/http"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"go.uber.org/zap"
)

const REALIPHEADER = "X-Real-IP"

// NewIPCheckMiddleware returns a middleware that verifies trusted subnet ip header for requests and responses.
func NewIPCheckMiddleware(cfg *config.ServerConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.TrustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			realIP := r.Header.Get(REALIPHEADER)
			if realIP == "" {
				logger.Log.Error("X-Real-IP header is missing")
				http.Error(w, "X-Real-IP header is required", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(realIP)
			if ip == nil {
				logger.Log.Error("Invalid IP address", zap.String("ip", realIP))
				http.Error(w, "Invalid IP address", http.StatusForbidden)
				return
			}

			_, subnet, err := net.ParseCIDR(cfg.TrustedSubnet)
			if err != nil {
				logger.Log.Error("Invalid trusted subnet", zap.Error(err))
				http.Error(w, "Server configuration error", http.StatusInternalServerError)
				return
			}

			if !subnet.Contains(ip) {
				logger.Log.Error("IP not in trusted subnet",
					zap.String("ip", realIP),
					zap.String("subnet", cfg.TrustedSubnet))
				http.Error(w, "Access denied: IP not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
