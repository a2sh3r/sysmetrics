package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewIPCheckMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		realIP         string
		expectedStatus int
		description    string
	}{
		{
			name:           "empty trusted subnet allows all",
			trustedSubnet:  "",
			realIP:         "192.168.1.100",
			expectedStatus: http.StatusOK,
			description:    "when trusted subnet is empty, all requests should be allowed",
		},
		{
			name:           "missing X-Real-IP header",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "",
			expectedStatus: http.StatusForbidden,
			description:    "when trusted subnet is configured, X-Real-IP header is required",
		},
		{
			name:           "IP in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "192.168.1.100",
			expectedStatus: http.StatusOK,
			description:    "IP address in trusted subnet should be allowed",
		},
		{
			name:           "IP not in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "192.168.2.100",
			expectedStatus: http.StatusForbidden,
			description:    "IP address not in trusted subnet should be denied",
		},
		{
			name:           "invalid IP address",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "invalid-ip",
			expectedStatus: http.StatusForbidden,
			description:    "invalid IP address should be denied",
		},
		{
			name:           "IPv6 in IPv4 subnet",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "2001:db8::1",
			expectedStatus: http.StatusForbidden,
			description:    "IPv6 address should not match IPv4 subnet",
		},
		{
			name:           "IPv6 in IPv6 subnet",
			trustedSubnet:  "2001:db8::/64",
			realIP:         "2001:db8::1",
			expectedStatus: http.StatusOK,
			description:    "IPv6 address in IPv6 subnet should be allowed",
		},
		{
			name:           "boundary IP in subnet",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "192.168.1.255",
			expectedStatus: http.StatusOK,
			description:    "boundary IP should be allowed",
		},
		{
			name:           "boundary IP outside subnet",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "192.168.2.0",
			expectedStatus: http.StatusForbidden,
			description:    "boundary IP outside subnet should be denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				TrustedSubnet: tt.trustedSubnet,
			}

			middleware := NewIPCheckMiddleware(cfg)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.realIP != "" {
				req.Header.Set(REALIPHEADER, tt.realIP)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code, tt.description)
		})
	}
}

func TestNewIPCheckMiddlewareWithInvalidSubnet(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		realIP         string
		expectedStatus int
	}{
		{
			name:           "invalid CIDR format",
			trustedSubnet:  "invalid-subnet",
			realIP:         "192.168.1.100",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid IP range",
			trustedSubnet:  "256.256.256.256/24",
			realIP:         "192.168.1.100",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				TrustedSubnet: tt.trustedSubnet,
			}

			middleware := NewIPCheckMiddleware(cfg)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set(REALIPHEADER, tt.realIP)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestNewIPCheckMiddleware_Integration(t *testing.T) {
	cfg := &config.ServerConfig{
		TrustedSubnet: "10.0.0.0/8",
	}

	middleware := NewIPCheckMiddleware(cfg)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("success")); err != nil {
			log.Printf("failed to write resp.Body: %v", err)
		}
	}))

	req := httptest.NewRequest("POST", "/updates/", nil)
	req.Header.Set(REALIPHEADER, "10.1.2.3")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())

	req = httptest.NewRequest("POST", "/updates/", nil)
	req.Header.Set(REALIPHEADER, "192.168.1.1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
