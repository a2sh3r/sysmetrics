package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
