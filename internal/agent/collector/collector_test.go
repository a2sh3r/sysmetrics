package collector

import (
	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollector_CollectMetrics(t *testing.T) {
	type fields struct {
		pollCount int64
	}
	tests := []struct {
		name   string
		fields fields
		want   *metrics.Metrics
	}{
		{
			name: "Test #1 collect metrics",
			fields: fields{
				pollCount: 0,
			},
			want: &metrics.Metrics{
				PollCount: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{
				pollCount: tt.fields.pollCount,
			}
			got := c.CollectMetrics()

			assert.NotNil(t, got)
			assert.Equal(t, tt.want.PollCount, got.PollCount)
		})
	}
}

func TestNewCollector(t *testing.T) {
	tests := []struct {
		name string
		want *Collector
	}{
		{
			name: "Test #1 new valid collector",
			want: &Collector{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCollector()
			assert.NotNil(t, got)
			assert.IsType(t, tt.want, got)
		})
	}
}
