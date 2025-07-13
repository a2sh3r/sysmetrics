package memstorage

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

func TestMemStorage_GetMetric(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		metrics map[string]repositories.Metric
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
			name: "Test #1 get existing metric",
			fields: fields{
				metrics: map[string]repositories.Metric{
					"test": {Type: constants.MetricTypeGauge, Value: 123.45},
				},
			},
			args: args{
				name: "test",
			},
			want:    repositories.Metric{Type: constants.MetricTypeGauge, Value: 123.45},
			wantErr: false,
		},
		{
			name: "Test #2 get non-existent metric",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				name: "non-existent",
			},
			want:    repositories.Metric{},
			wantErr: true,
		},
		{
			name: "Test #3 get metric from nil storage",
			fields: fields{
				metrics: nil,
			},
			args: args{
				name: "test",
			},
			want:    repositories.Metric{},
			wantErr: true,
		},
		{
			name: "Test #4 get metric from nil MemStorage",
			fields: fields{
				metrics: nil,
			},
			args: args{
				name: "test",
			},
			want:    repositories.Metric{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			got, err := ms.GetMetric(ctx, tt.args.name)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMemStorage_GetMetrics(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		metrics map[string]repositories.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]repositories.Metric
		wantErr bool
	}{
		{
			name: "Test #1 get existing metric",
			fields: fields{
				metrics: map[string]repositories.Metric{
					"test_gauge":   {Type: constants.MetricTypeGauge, Value: 123.45},
					"test_counter": {Type: constants.MetricTypeCounter, Value: int64(123)},
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
			ms := &MemStorage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			got, err := ms.GetMetrics(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMemStorage_UpdateMetric(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		metrics map[string]repositories.Metric
	}
	type args struct {
		name   string
		metric repositories.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test #1 update existing metric",
			fields: fields{
				metrics: map[string]repositories.Metric{
					"test": {Type: constants.MetricTypeCounter, Value: int64(10)},
				},
			},
			args: args{
				name: "test",
				metric: repositories.Metric{
					Type:  constants.MetricTypeCounter,
					Value: int64(20),
				},
			},
			wantErr: false,
		},
		{
			name: "Test #2 update non-existent metric",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				name: "new",
				metric: repositories.Metric{
					Type:  constants.MetricTypeGauge,
					Value: 123.45,
				},
			},
			wantErr: false,
		},
		{
			name: "Test #3 update metric with invalid type",
			fields: fields{
				metrics: map[string]repositories.Metric{
					"test": {Type: constants.MetricTypeCounter, Value: int64(10)},
				},
			},
			args: args{
				name: "test",
				metric: repositories.Metric{
					Type:  "invalid",
					Value: int64(20),
				},
			},
			wantErr: true,
		},
		{
			name: "Test #4 update metric with invalid name",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				name: "",
				metric: repositories.Metric{
					Type:  constants.MetricTypeGauge,
					Value: 123.45,
				},
			},
			wantErr: true,
		},
		{
			name: "Test #5 update metric from nil storage",
			fields: fields{
				metrics: nil,
			},
			args: args{
				name: "test",
				metric: repositories.Metric{
					Type:  constants.MetricTypeGauge,
					Value: 123.45,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			err := ms.UpdateMetric(ctx, tt.args.name, tt.args.metric)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMemStorage_updateCounterMetric(t *testing.T) {
	type fields struct {
		metrics map[string]repositories.Metric
	}
	type args struct {
		existingMetric *repositories.Metric
		newMetric      repositories.Metric
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
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				existingMetric: &repositories.Metric{
					Type:  constants.MetricTypeCounter,
					Value: int64(10),
				},
				newMetric: repositories.Metric{
					Type:  constants.MetricTypeCounter,
					Value: int64(20),
				},
			},
			wantErr: false,
		},
		{
			name: "Test #2 update counter metric with invalid type",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				existingMetric: &repositories.Metric{
					Type:  constants.MetricTypeCounter,
					Value: int64(10),
				},
				newMetric: repositories.Metric{
					Type:  "invalid",
					Value: int64(20),
				},
			},
			wantErr: true,
		},
		{
			name: "Test #3 update counter metric with invalid value type",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				existingMetric: &repositories.Metric{
					Type:  constants.MetricTypeCounter,
					Value: int64(10),
				},
				newMetric: repositories.Metric{
					Type:  constants.MetricTypeCounter,
					Value: "invalid",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			err := ms.updateCounterMetric(tt.args.existingMetric, tt.args.newMetric)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMemStorage_updateGaugeMetric(t *testing.T) {
	type fields struct {
		metrics map[string]repositories.Metric
	}
	type args struct {
		existingMetric *repositories.Metric
		newMetric      repositories.Metric
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
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				existingMetric: &repositories.Metric{
					Type:  constants.MetricTypeGauge,
					Value: 10.5,
				},
				newMetric: repositories.Metric{
					Type:  constants.MetricTypeGauge,
					Value: 20.5,
				},
			},
			wantErr: false,
		},
		{
			name: "Test #2 update gauge metric with invalid type",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				existingMetric: &repositories.Metric{
					Type:  constants.MetricTypeGauge,
					Value: 10.5,
				},
				newMetric: repositories.Metric{
					Type:  "invalid",
					Value: 20.5,
				},
			},
			wantErr: true,
		},
		{
			name: "Test #3 update gauge metric with invalid value type",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				existingMetric: &repositories.Metric{
					Type:  constants.MetricTypeGauge,
					Value: 10.5,
				},
				newMetric: repositories.Metric{
					Type:  constants.MetricTypeGauge,
					Value: "invalid",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			err := ms.updateGaugeMetric(tt.args.existingMetric, tt.args.newMetric)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MemStorage
	}{
		{
			name: "Test #1 create new MemStorage",
			want: &MemStorage{
				metrics: make(map[string]repositories.Metric),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMemStorage()
			assert.NotNil(t, got)
			assert.NotNil(t, got.metrics)
		})
	}
}

func TestMemStorage_UpdateMetricsBatch(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		metrics map[string]repositories.Metric
	}
	type args struct {
		metrics map[string]repositories.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]repositories.Metric
		wantErr bool
	}{
		{
			name: "Batch update: add new metrics",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				metrics: map[string]repositories.Metric{
					"gauge1":   {Type: constants.MetricTypeGauge, Value: 1.23},
					"counter1": {Type: constants.MetricTypeCounter, Value: int64(10)},
				},
			},
			want: map[string]repositories.Metric{
				"gauge1":   {Type: constants.MetricTypeGauge, Value: 1.23},
				"counter1": {Type: constants.MetricTypeCounter, Value: int64(10)},
			},
			wantErr: false,
		},
		{
			name: "Batch update: update existing counter",
			fields: fields{
				metrics: map[string]repositories.Metric{
					"counter1": {Type: constants.MetricTypeCounter, Value: int64(5)},
				},
			},
			args: args{
				metrics: map[string]repositories.Metric{
					"counter1": {Type: constants.MetricTypeCounter, Value: int64(7)},
				},
			},
			want: map[string]repositories.Metric{
				"counter1": {Type: constants.MetricTypeCounter, Value: int64(12)},
			},
			wantErr: false,
		},
		{
			name: "Batch update: update existing gauge",
			fields: fields{
				metrics: map[string]repositories.Metric{
					"gauge1": {Type: constants.MetricTypeGauge, Value: 2.34},
				},
			},
			args: args{
				metrics: map[string]repositories.Metric{
					"gauge1": {Type: constants.MetricTypeGauge, Value: 3.45},
				},
			},
			want: map[string]repositories.Metric{
				"gauge1": {Type: constants.MetricTypeGauge, Value: 3.45},
			},
			wantErr: false,
		},
		{
			name: "Batch update: invalid metric type",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				metrics: map[string]repositories.Metric{
					"bad": {Type: "invalid", Value: 1},
				},
			},
			want:    map[string]repositories.Metric{},
			wantErr: true,
		},
		{
			name: "Batch update: empty metric name",
			fields: fields{
				metrics: map[string]repositories.Metric{},
			},
			args: args{
				metrics: map[string]repositories.Metric{
					"": {Type: constants.MetricTypeGauge, Value: 1.23},
				},
			},
			want:    map[string]repositories.Metric{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				metrics: tt.fields.metrics,
				mu:      sync.RWMutex{},
			}
			err := ms.UpdateMetricsBatch(ctx, tt.args.metrics)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, ms.metrics)
			}
		})
	}
}

func BenchmarkUpdateMetric(b *testing.B) {
	ms := NewMemStorage()
	ctx := context.Background()
	metric := repositories.Metric{Type: "gauge", Value: float64(42)}
	for i := 0; i < b.N; i++ {
		_ = ms.UpdateMetric(ctx, "test_metric", metric)
	}
}
