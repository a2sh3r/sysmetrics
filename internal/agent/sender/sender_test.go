package sender

import (
	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSender(t *testing.T) {
	type args struct {
		serverAddress string
	}
	tests := []struct {
		name string
		args args
		want *Sender
	}{
		{
			name: "Test #1 create valid sender",
			args: args{
				serverAddress: "http://localhost:8080",
			},
			want: &Sender{
				serverAddress: "http://localhost:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSender(tt.args.serverAddress)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSender_SendMetrics(t *testing.T) {
	type fields struct {
		serverAddress string
	}
	type args struct {
		metricsBatch []*metrics.Metrics
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test #1 send valid metrics",
			fields: fields{
				serverAddress: "http://localhost:8080",
			},
			args: args{
				metricsBatch: []*metrics.Metrics{
					{
						PollCount: int64(1),
						HeapAlloc: 11.11,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test #2 send empty metrics",
			fields: fields{
				serverAddress: "http://localhost:8080",
			},
			args: args{
				metricsBatch: []*metrics.Metrics{},
			},
			wantErr: true,
		},
		{
			name: "Test #3 send metrics to invalid",
			fields: fields{
				serverAddress: "http://localhost111:8080",
			},
			args: args{
				metricsBatch: []*metrics.Metrics{
					{
						PollCount: int64(1),
						HeapAlloc: 11.11,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			if tt.name != "Test #3 send metrics to invalid" {
				tt.fields.serverAddress = srv.URL
			}

			s := &Sender{
				serverAddress: tt.fields.serverAddress,
			}
			err := s.SendMetrics(tt.args.metricsBatch)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSender_sendMetric(t *testing.T) {
	type fields struct {
		serverAddress string
	}
	type args struct {
		metricType string
		metricName string
		value      interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test #1 send counter metric",
			fields: fields{
				serverAddress: "http://localhost:8080",
			},
			args: args{
				metricType: constants.MetricTypeCounter,
				metricName: "PollCount",
				value:      int64(1),
			},
			wantErr: false,
		},
		{
			name: "Test #2 send invalid counter metric",
			fields: fields{
				serverAddress: "http://localhost:8080",
			},
			args: args{
				metricType: "counter",
				metricName: "PollCount",
				value:      11.1,
			},
			wantErr: true,
		},
		{
			name: "Test #3 send valid gauge metric",
			fields: fields{
				serverAddress: "http://localhost:8080",
			},
			args: args{
				metricType: constants.MetricTypeGauge,
				metricName: "Alloc",
				value:      11.1,
			},
			wantErr: false,
		},
		{
			name: "Test #4 send unsupported metric type",
			fields: fields{
				serverAddress: "http://localhost:8080",
			},
			args: args{
				metricType: "test",
				metricName: "Alloc",
				value:      "test",
			},
			wantErr: true,
		},
		{
			name: "Test #5 send to invalid address",
			fields: fields{
				serverAddress: "http://localhoststs:8080",
			},
			args: args{
				metricType: constants.MetricTypeGauge,
				metricName: "Alloc",
				value:      11.0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			if tt.name != "Test #5 send to invalid address" {
				tt.fields.serverAddress = srv.URL
			}

			s := &Sender{
				serverAddress: tt.fields.serverAddress,
			}
			err := s.sendMetric(tt.args.metricType, tt.args.metricName, tt.args.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
