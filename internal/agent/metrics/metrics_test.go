package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	tests := []struct {
		name string
		want *Metrics
	}{
		{
			name: "Test #1 create new config",
			want: &Metrics{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetrics()
			assert.NotNil(t, got)
			assert.IsType(t, tt.want, got)
		})
	}
}

func BenchmarkNewMetrics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewMetrics()
	}
}
