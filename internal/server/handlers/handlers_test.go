package handlers

import (
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_UpdateMetric(t *testing.T) {
	type fields struct {
		service *metric.Service
	}
	type args struct {
		method string
		url    string
	}
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Test #1 update gauge metric",
			fields: fields{
				service: metric.NewService(&mockRepo{}),
			},
			args: args{
				method: http.MethodPost,
				url:    "/update/gauge/test/123.45",
			},
			want: want{
				code:        http.StatusOK,
				response:    "Metric test is updated successfully with value 123.45",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #2 update counter metric",
			fields: fields{
				service: metric.NewService(&mockRepo{}),
			},
			args: args{
				method: http.MethodPost,
				url:    "/update/counter/test/123",
			},
			want: want{
				code:        http.StatusOK,
				response:    "Metric test is updated successfully with value 123",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #3 invalid method",
			fields: fields{
				service: metric.NewService(&mockRepo{}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/update/gauge/test/123.45",
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "Method not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #4 invalid URL format",
			fields: fields{
				service: metric.NewService(&mockRepo{}),
			},
			args: args{
				method: http.MethodPost,
				url:    "/invalid",
			},
			want: want{
				code:        http.StatusNotFound,
				response:    "Invalid URL format\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #5 invalid metric type",
			fields: fields{
				service: metric.NewService(&mockRepo{}),
			},
			args: args{
				method: http.MethodPost,
				url:    "/update/invalid/test/123",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "Failed to update metric: invalid metric type: invalid\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #6 invalid metric value",
			fields: fields{
				service: metric.NewService(&mockRepo{}),
			},
			args: args{
				method: http.MethodPost,
				url:    "/update/gauge/test/invalid",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "Failed to update metric: strconv.ParseFloat: parsing \"invalid\": invalid syntax\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
			}

			ts := httptest.NewServer(NewRouter(h))
			defer ts.Close()

			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, nil)
			require.NoError(t, err)

			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					return
				}
			}(res.Body)

			assert.Equal(t, tt.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.response, string(resBody))
		})
	}
}

func TestHandler_GetMetric(t *testing.T) {
	type fields struct {
		service *metric.Service
	}
	type args struct {
		method string
		url    string
	}
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Test #1 get gauge metric",
			fields: fields{
				service: metric.NewService(&mockRepo{
					metrics: map[string]repositories.Metric{
						"test_gauge": {
							Type:  "gauge",
							Value: 123.123,
						},
					},
				}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/value/gauge/test_gauge",
			},
			want: want{
				code:        http.StatusOK,
				response:    "123.123\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #2 get counter metric",
			fields: fields{
				service: metric.NewService(&mockRepo{
					metrics: map[string]repositories.Metric{
						"test_counter": {
							Type:  "counter",
							Value: int64(123),
						},
					},
				}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/value/counter/test_counter",
			},
			want: want{
				code:        http.StatusOK,
				response:    "123\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #3 get invalid type metric",
			fields: fields{
				service: metric.NewService(&mockRepo{
					metrics: map[string]repositories.Metric{
						"test_counter": {
							Type:  "counter",
							Value: int64(123),
						},
					},
				}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/value/gauge/test_counter",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "Got metric, but its type differs from the requested one\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #3 get invalid type metric",
			fields: fields{
				service: metric.NewService(&mockRepo{
					metrics: map[string]repositories.Metric{
						"test_counter": {
							Type:  "counter",
							Value: int64(123),
						},
					},
				}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/value/gauge/test_counter",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "Got metric, but its type differs from the requested one\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #4 failed to get metric",
			fields: fields{
				service: metric.NewService(&mockRepo{}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/value/gauge/test_counter",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "Failed to get metric: metric test_counter not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #5 metric is nil",
			fields: fields{
				service: metric.NewService(&mockRepo{
					metrics: map[string]repositories.Metric{
						"test_nil": {
							Type:  "gauge",
							Value: nil,
						},
					},
				}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/value/gauge/test_nil",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "Metric value is nil\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
			}

			ts := httptest.NewServer(NewRouter(h))
			defer ts.Close()

			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, nil)
			require.NoError(t, err)

			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					return
				}
			}(res.Body)

			assert.Equal(t, tt.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.response, string(resBody))
		})
	}
}

func TestHandler_GetMetrics(t *testing.T) {
	type fields struct {
		service *metric.Service
	}
	type args struct {
		method string
		url    string
	}
	type want struct {
		code        int
		response    []string
		contentType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Test #1 get gauge metric",
			fields: fields{
				service: metric.NewService(&mockRepo{
					metrics: map[string]repositories.Metric{
						"test_gauge": {
							Type:  "gauge",
							Value: 123.123,
						},
						"test_counter": {
							Type:  "counter",
							Value: int64(123),
						},
					},
				}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/",
			},
			want: want{
				code: http.StatusOK,
				response: []string{
					"test_gauge 123.123",
					"test_counter 123",
				},
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #2 get invalid metric",
			fields: fields{
				service: metric.NewService(&mockRepo{
					metrics: map[string]repositories.Metric{
						"test_invalid": {
							Type:  "gauge",
							Value: "String",
						},
					},
				}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/",
			},
			want: want{
				code: http.StatusInternalServerError,
				response: []string{
					"Unsupported metric value type",
				},
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
			}

			ts := httptest.NewServer(NewRouter(h))
			defer ts.Close()

			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, nil)
			require.NoError(t, err)

			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					return
				}
			}(res.Body)

			assert.Equal(t, tt.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			actualLines := strings.Split(strings.TrimSpace(string(resBody)), "\n")

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			for _, expectedLine := range tt.want.response {
				assert.Contains(t, actualLines, expectedLine)
			}
			assert.Equal(t, len(tt.want.response), len(actualLines))
		})
	}
}

func TestNewHandler(t *testing.T) {
	type args struct {
		service *metric.Service
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "Test #1 create handler with valid service",
			args: args{
				service: metric.NewService(&mockRepo{}),
			},
			want: &Handler{
				service: metric.NewService(&mockRepo{}),
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
