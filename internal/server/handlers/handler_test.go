package handlers

import (
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHandler(t *testing.T) {
	type args struct {
		service *services.Service
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "Test #1 create handler with valid service",
			args: args{
				service: services.NewService(&mockRepo{}),
			},
			want: &Handler{
				service: services.NewService(&mockRepo{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHandler(tt.args.service)
			assert.NotNil(t, got)
			assert.Equal(t, got, tt.want)
		})
	}
}

type mockRepo struct {
	metrics     map[string]repositories.Metric
	errOnUpdate bool
	errOnGet    bool
}

func (m *mockRepo) GetMetric(name string) (repositories.Metric, error) {
	if m.errOnGet {
		return repositories.Metric{}, fmt.Errorf("mock get error")
	}
	getMetric, ok := m.metrics[name]
	if !ok {
		return repositories.Metric{}, fmt.Errorf("metric %s not found", name)
	}
	return getMetric, nil
}

func (m *mockRepo) GetMetrics() (map[string]repositories.Metric, error) {
	if m.errOnGet {
		return map[string]repositories.Metric{}, fmt.Errorf("mock get error")
	}
	return m.metrics, nil
}

func (m *mockRepo) SaveMetric(name string, value interface{}, metricType string) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error with %v, %v, %v", name, value, metricType)
	}
	return nil
}
