package repositories

import (
	"context"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockStorage struct {
	metrics     map[string]Metric
	errOnUpdate bool
	errOnGet    bool
}

func (m *mockStorage) SaveMetric(_ context.Context, name string, value interface{}, metricType string) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error")
	}
	if m.metrics == nil {
		m.metrics = make(map[string]Metric)
	}
	m.metrics[name] = Metric{
		Type:  metricType,
		Value: value,
	}
	return nil
}

func (m *mockStorage) UpdateMetric(_ context.Context, name string, metric Metric) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error")
	}
	if m.metrics == nil {
		m.metrics = make(map[string]Metric)
	}
	m.metrics[name] = metric
	return nil
}

func (m *mockStorage) GetMetric(_ context.Context, name string) (Metric, error) {
	if m.errOnGet {
		return Metric{}, fmt.Errorf("mock get error")
	}
	metric, ok := m.metrics[name]
	if !ok {
		return Metric{}, fmt.Errorf("metric %s not found", name)
	}
	return metric, nil
}

func (m *mockStorage) GetMetrics(_ context.Context) (map[string]Metric, error) {
	if m.errOnGet {
		return map[string]Metric{}, fmt.Errorf("mock get error")
	}
	return m.metrics, nil
}

func (m *mockStorage) UpdateMetricsBatch(_ context.Context, metrics map[string]Metric) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error")
	}
	if m.metrics == nil {
		m.metrics = make(map[string]Metric)
	}
	for name, metric := range metrics {
		m.metrics[name] = metric
	}
	return nil
}

func TestMetricRepo_GetMetric(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		storage Storage
		metric  string
		want    Metric
		wantErr bool
	}{
		{
			name: "Test #1 get existing metric",
			storage: &mockStorage{
				metrics: map[string]Metric{
					"test": {
						Type:  constants.MetricTypeGauge,
						Value: 123.45,
					},
				},
			},
			metric: "test",
			want: Metric{
				Type:  constants.MetricTypeGauge,
				Value: 123.45,
			},
			wantErr: false,
		},
		{
			name:    "Test #2 get non-existent metric",
			storage: &mockStorage{},
			metric:  "test",
			want:    Metric{},
			wantErr: true,
		},
		{
			name: "Test #3 get metric with error",
			storage: &mockStorage{
				errOnGet: true,
			},
			metric:  "test",
			want:    Metric{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MetricRepo{
				storage: tt.storage,
			}
			got, err := r.GetMetric(ctx, tt.metric)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMetricRepo_GetMetrics(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		storage Storage
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]Metric
		wantErr bool
	}{
		{
			name: "Test #1 get metrics",
			fields: fields{
				storage: &mockStorage{
					metrics: map[string]Metric{
						"test":  {Type: constants.MetricTypeGauge, Value: 123.45},
						"test2": {Type: constants.MetricTypeCounter, Value: 123},
					},
				},
			},
			want: map[string]Metric{
				"test":  {Type: constants.MetricTypeGauge, Value: 123.45},
				"test2": {Type: constants.MetricTypeCounter, Value: 123},
			},
			wantErr: false,
		},
		{
			name: "Test #2 get nil map",
			fields: fields{
				storage: &mockStorage{
					metrics: nil,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MetricRepo{
				storage: tt.fields.storage,
			}
			got, err := r.GetMetrics(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMetricRepo_SaveMetric(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		storage Storage
	}
	type args struct {
		name       string
		value      interface{}
		metricType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test #1 save metric",
			fields: fields{
				storage: NewMockStorage(),
			},
			args: args{
				name:       "test",
				value:      123.45,
				metricType: constants.MetricTypeGauge,
			},
			wantErr: false,
		},
		{
			name: "Test #2 save metric with error",
			fields: fields{
				storage: &MockStorage{errOnUpdate: true},
			},
			args: args{
				name:       "test",
				value:      123.45,
				metricType: constants.MetricTypeGauge,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MetricRepo{
				storage: tt.fields.storage,
			}

			err := r.SaveMetric(ctx, tt.args.name, tt.args.value, tt.args.metricType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewMetricRepo(t *testing.T) {
	type args struct {
		storage Storage
	}
	tests := []struct {
		name string
		args args
		want *MetricRepo
	}{
		{
			name: "Test #1 create metric repo",
			args: args{
				storage: NewMockStorage(),
			},
			want: &MetricRepo{
				storage: NewMockStorage(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetricRepo(tt.args.storage)
			assert.NotNil(t, got)
			assert.Equal(t, got, tt.want)
		})
	}
}

type MockStorage struct {
	metrics     map[string]Metric
	errOnUpdate bool
	errOnGet    bool
}

func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

func (m *MockStorage) UpdateMetric(_ context.Context, metricName string, metric Metric) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error")
	}
	if m.metrics == nil {
		m.metrics = make(map[string]Metric)
	}
	m.metrics[metricName] = metric
	return nil
}

func (m *MockStorage) GetMetric(_ context.Context, metricName string) (Metric, error) {
	if m.errOnGet {
		return Metric{}, fmt.Errorf("mock get error")
	}
	metric, ok := m.metrics[metricName]
	if !ok {
		return Metric{}, fmt.Errorf("metric %s not found", metricName)
	}
	return metric, nil
}

func (m *MockStorage) GetMetrics(_ context.Context) (map[string]Metric, error) {
	if m.errOnGet {
		return map[string]Metric{}, fmt.Errorf("mock get error")
	}
	return m.metrics, nil
}

func (m *MockStorage) UpdateMetricsBatch(_ context.Context, metrics map[string]Metric) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error")
	}
	if m.metrics == nil {
		m.metrics = make(map[string]Metric)
	}
	for name, metric := range metrics {
		m.metrics[name] = metric
	}
	return nil
}
