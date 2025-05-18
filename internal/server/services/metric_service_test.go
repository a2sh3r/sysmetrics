package services

import (
	"context"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewService(t *testing.T) {
	type args struct {
		repo MetricRepository
	}
	tests := []struct {
		name string
		args args
		want *Service
	}{
		{
			name: "Test #1 create service with valid repo",
			args: args{
				repo: &mockRepo{},
			},
			want: &Service{
				repo: &mockRepo{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewService(tt.args.repo)
			assert.NotNil(t, got)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestService_UpdateCounterMetric(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		repo MetricRepository
	}
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test #1 update counter metric",
			fields: fields{
				repo: &mockRepo{},
			},
			args: args{
				name:  "test",
				value: 123,
			},
			wantErr: false,
		},
		{
			name: "Test #2 update counter metric with error",
			fields: fields{
				repo: &mockRepo{errOnUpdate: true},
			},
			args: args{
				name:  "test",
				value: 123,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				repo: tt.fields.repo,
			}
			err := s.UpdateCounterMetric(ctx, tt.args.name, tt.args.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_UpdateGaugeMetric(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		repo MetricRepository
	}
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test #1 update gauge metric",
			fields: fields{
				repo: &mockRepo{},
			},
			args: args{
				name:  "test",
				value: 123.45,
			},
			wantErr: false,
		},
		{
			name: "Test #2 update gauge metric with error",
			fields: fields{
				repo: &mockRepo{errOnUpdate: true},
			},
			args: args{
				name:  "test",
				value: 123.45,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				repo: tt.fields.repo,
			}
			err := s.UpdateGaugeMetric(ctx, tt.args.name, tt.args.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetMetrics(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		repo MetricRepository
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]repositories.Metric
		wantErr bool
	}{
		{
			name: "Test #1 get metrics",
			fields: fields{
				repo: &mockRepo{
					metrics: map[string]repositories.Metric{
						"test_gauge": {
							Type:  constants.MetricTypeGauge,
							Value: 123.45,
						},
						"test_counter": {
							Type:  constants.MetricTypeCounter,
							Value: int64(123),
						},
					},
				},
			},
			want: map[string]repositories.Metric{
				"test_gauge": {
					Type:  constants.MetricTypeGauge,
					Value: 123.45,
				},
				"test_counter": {
					Type:  constants.MetricTypeCounter,
					Value: int64(123),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				repo: tt.fields.repo,
			}
			got, err := s.GetMetrics(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_GetMetric(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		repo MetricRepository
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    repositories.Metric
		wantErr bool
	}{
		{
			name: "Test #1 get metrics",
			fields: fields{
				repo: &mockRepo{
					metrics: map[string]repositories.Metric{
						"test_gauge": {
							Type:  constants.MetricTypeGauge,
							Value: 123.45,
						},
						"test_counter": {
							Type:  constants.MetricTypeCounter,
							Value: int64(123),
						},
					},
				},
			},
			args: args{
				name: "test_gauge",
			},
			want: repositories.Metric{
				Type:  constants.MetricTypeGauge,
				Value: 123.45,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				repo: tt.fields.repo,
			}
			got, err := s.GetMetric(ctx, tt.args.name)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

type mockRepo struct {
	metrics     map[string]repositories.Metric
	errOnUpdate bool
	errOnGet    bool
}

func (m *mockRepo) GetMetric(_ context.Context, name string) (repositories.Metric, error) {
	if m.errOnGet {
		return repositories.Metric{}, fmt.Errorf("mock get error")
	}
	metric, ok := m.metrics[name]
	if !ok {
		return repositories.Metric{}, fmt.Errorf("metric %s not found", name)
	}
	return metric, nil
}

func (m *mockRepo) GetMetrics(_ context.Context) (map[string]repositories.Metric, error) {
	if m.errOnGet {
		return map[string]repositories.Metric{}, fmt.Errorf("mock get error")
	}
	return m.metrics, nil
}

func (m *mockRepo) SaveMetric(_ context.Context, name string, value interface{}, metricType string) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error with %v, %v, %v", name, value, metricType)
	}
	return nil
}

func (m *mockRepo) UpdateMetricsBatch(_ context.Context, metrics map[string]repositories.Metric) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error with %v", metrics)
	}
	return nil
}
