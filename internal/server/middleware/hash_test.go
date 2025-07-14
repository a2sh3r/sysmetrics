package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/hash"
)

func TestHashMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		secretKey          string
		requestBody        string
		requestHash        string
		expectedStatus     int
		expectResponseHash bool
	}{
		{
			name:               "Test #1 no key",
			secretKey:          "",
			requestBody:        "test data",
			requestHash:        "",
			expectedStatus:     http.StatusOK,
			expectResponseHash: false,
		},
		{
			name:               "Test #2 valid hash",
			secretKey:          "test key",
			requestBody:        "test data",
			requestHash:        hash.CalculateHash("test data", "test key"),
			expectedStatus:     http.StatusOK,
			expectResponseHash: true,
		},
		{
			name:               "Test #3 invalid hash",
			secretKey:          "test key",
			requestBody:        "test data",
			requestHash:        "invalid hash",
			expectedStatus:     http.StatusBadRequest,
			expectResponseHash: false,
		},
		{
			name:               "Test #4 missing hash",
			secretKey:          "test key",
			requestBody:        "test data",
			requestHash:        "",
			expectedStatus:     http.StatusOK,
			expectResponseHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				SecretKey: tt.secretKey,
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("response"))
			})

			middleware := NewHashMiddleware(cfg)
			server := httptest.NewServer(middleware(handler))
			defer server.Close()

			req, err := http.NewRequest("POST", server.URL, bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}

			if tt.requestHash != "" {
				req.Header.Set(HashHeader, tt.requestHash)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("failed to close resp.Body: %v", err)
				}
			}()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			responseHash := resp.Header.Get(HashHeader)
			if tt.expectResponseHash {
				if responseHash == "" {
					t.Error("expected hash in response headers")
				} else {
					expectedHash := hash.CalculateHash("response", tt.secretKey)
					if responseHash != expectedHash {
						t.Errorf("expected response hash %s, got %s", expectedHash, responseHash)
					}
				}
			} else {
				if responseHash != "" {
					t.Error("unexpected hash in response headers")
				}
			}
		})
	}
}
