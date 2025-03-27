package repositories

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricRepo_GetMetric(t *testing.T) {
	type fields struct {
		storage Storage
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Metric
		wantErr bool
	}{
		{
			name: "Test #1 get existing metric",
			fields: fields{
				storage: &MockStorage{
					metrics: map[string]Metric{
						"test": {Type: "gauge", Value: 123.45},
					},
				},
			},
			args: args{
				name: "test",
			},
			want:    Metric{Type: "gauge", Value: 123.45},
			wantErr: false,
		},
		{
			name: "Test #2 get non-existent metric",
			fields: fields{
				storage: NewMockStorage(),
			},
			args: args{
				name: "non-existent",
			},
			want:    Metric{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MetricRepo{
				storage: tt.fields.storage,
			}
			got, err := r.GetMetric(tt.args.name)

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
				metricType: "gauge",
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
				metricType: "gauge",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MetricRepo{
				storage: tt.fields.storage,
			}

			err := r.SaveMetric(tt.args.name, tt.args.value, tt.args.metricType)
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
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		metrics: make(map[string]Metric),
	}
}

func (m *MockStorage) UpdateMetric(name string, metric Metric) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error")
	}
	m.metrics[name] = metric
	return nil
}

func (m *MockStorage) GetMetric(name string) (Metric, error) {
	metric, ok := m.metrics[name]
	if !ok {
		return Metric{}, fmt.Errorf("mock update error")
	}
	return metric, nil
}
