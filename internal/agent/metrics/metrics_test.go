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

func TestMetrics_UpdateSystemMetrics(t *testing.T) {
	tests := []struct {
		name string
		prepare func(m *Metrics)
		check func(t *testing.T, m *Metrics, err error)
	}{
		{
			name: "basic call",
			prepare: func(m *Metrics) {},
			check: func(t *testing.T, m *Metrics, err error) {
				assert.NoError(t, err)
				assert.NotZero(t, m.TotalMemory)
				assert.NotZero(t, m.FreeMemory)
				assert.NotNil(t, m.CPUUtilization)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics()
			if tt.prepare != nil {
				tt.prepare(m)
			}
			err := m.UpdateSystemMetrics()
			if tt.check != nil {
				tt.check(t, m, err)
			}
		})
	}
}
