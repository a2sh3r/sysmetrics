package agent

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricsWorker(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit int64
	}{
		{"rate 1", 1},
		{"rate 5", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewMetricsWorker(tt.rateLimit, func(m *metrics.Metrics) error { return nil })
			assert.NotNil(t, w)
			assert.Equal(t, tt.rateLimit, w.rateLimit)
			assert.NotNil(t, w.metricsChan)
		})
	}
}

func TestMetricsWorker_Start_Send_Stop(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit int64
		sendCount int
	}{
		{"single metric", 1, 1},
		{"multiple metrics", 2, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var processed int32
			w := NewMetricsWorker(tt.rateLimit, func(m *metrics.Metrics) error {
				atomic.AddInt32(&processed, 1)
				return nil
			})
			ctx, cancel := context.WithCancel(context.Background())
			go w.Start(ctx)
			for i := 0; i < tt.sendCount; i++ {
				w.SendMetrics(metrics.NewMetrics())
			}
			time.Sleep(50 * time.Millisecond)
			cancel()
			w.Stop()
			assert.GreaterOrEqual(t, processed, int32(tt.sendCount))
		})
	}
} 