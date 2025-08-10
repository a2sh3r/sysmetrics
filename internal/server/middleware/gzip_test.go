package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGzipWriter_Write(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantErr bool
	}{
		{"basic write", "hello", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)
			w := &gzipWriter{writer: gz, ResponseWriter: httptest.NewRecorder()}
			_, err := w.Write([]byte(tt.input))
			_ = gz.Close()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, buf.Len() > 0)
			}
		})
	}
}

func TestNewGzipMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		contentEncoding string
		body           string
		wantGzip       bool
		wantStatus     int
	}{
		{"gzip response", "gzip", "", "hello", true, 200},
		{"no gzip", "", "", "hello", false, 200},
		{"gzip request", "", "gzip", "hello", false, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewGzipMiddleware()
			req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(tt.body))
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				_, _ = gz.Write([]byte(tt.body))
				_ = gz.Close()
				req.Body = io.NopCloser(&buf)
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			rw := httptest.NewRecorder()
			h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				_, _ = w.Write([]byte("ok"))
			}))
			h.ServeHTTP(rw, req)
			if tt.wantGzip {
				assert.Equal(t, "gzip", rw.Header().Get("Content-Encoding"))
			} else {
				assert.NotEqual(t, "gzip", rw.Header().Get("Content-Encoding"))
			}
			assert.Equal(t, tt.wantStatus, rw.Code)
		})
	}
} 