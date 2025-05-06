package sender

import (
	"context"
	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSender_SendMetrics(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name          string
		serverAddress string
		metricsBatch  []*metrics.Metrics
		wantErr       bool
	}{
		{
			name:          "valid metrics with PollCount and HeapAlloc",
			serverAddress: "http://localhost:8080",
			metricsBatch: []*metrics.Metrics{
				{
					PollCount: int64(1),
					HeapAlloc: 12345.67,
				},
			},
			wantErr: false,
		},
		{
			name:          "empty metrics batch",
			serverAddress: "http://localhost:8080",
			metricsBatch:  []*metrics.Metrics{},
			wantErr:       true,
		},
		{
			name:          "nil metrics batch",
			serverAddress: "http://localhost:8080",
			metricsBatch:  nil,
			wantErr:       true,
		},
		{
			name:          "invalid server address",
			serverAddress: "http://invalid-address",
			metricsBatch: []*metrics.Metrics{
				{
					PollCount: int64(1),
					HeapAlloc: 12345.67,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}
				if r.URL.Path != "/updates/" {
					t.Errorf("expected path /updates/, got %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			if tt.serverAddress == "http://localhost:8080" {
				tt.serverAddress = srv.URL
			}

			s := &Sender{
				serverAddress: tt.serverAddress,
				client:        &http.Client{Timeout: 5 * time.Second},
			}

			err := s.SendMetrics(ctx, tt.metricsBatch)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
