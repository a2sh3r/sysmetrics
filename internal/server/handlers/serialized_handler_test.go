package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a2sh3r/sysmetrics/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/models"
	"github.com/a2sh3r/sysmetrics/internal/server/middleware"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
)

func TestHandler_UpdateSerializedMetric(t *testing.T) {
	type args struct {
		body   models.Metrics
		method string
		url    string
	}
	sharedRepo := &mockRepo{
		metrics: make(map[string]repositories.Metric),
	}
	tests := []struct {
		name              string
		mockReaderService ReaderServiceInterface
		mockWriterService WriterServiceInterface
		args              args
		wantStatusCode    int
		wantContent       string
	}{
		{
			name:              "Test #1 valid gauge metric update",
			mockReaderService: services.NewService(sharedRepo),
			mockWriterService: services.NewService(sharedRepo),
			args: args{
				method: http.MethodPost,
				url:    "/update/",
				body: models.Metrics{
					ID:    "gauge_metric",
					MType: constants.MetricTypeGauge,
					Value: float64Ptr(123.12),
				},
			},
			wantStatusCode: http.StatusOK,
			wantContent:    `{"id":"gauge_metric","type":"gauge","value":123.12}`,
		},
		{
			name:              "Test #2 missing gauge value",
			mockReaderService: services.NewService(sharedRepo),
			mockWriterService: services.NewService(sharedRepo),
			args: args{
				method: http.MethodPost,
				url:    "/update/",
				body: models.Metrics{
					ID:    "gauge_metric",
					MType: constants.MetricTypeGauge,
				},
			},
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "Missing gauge value",
		},
		{
			name:              "Test #3 unsupported metric type",
			mockReaderService: services.NewService(sharedRepo),
			mockWriterService: services.NewService(sharedRepo),
			args: args{
				method: http.MethodPost,
				url:    "/update/",
				body: models.Metrics{
					ID:    "metric_unknown",
					MType: "unknown",
				},
			},
			wantStatusCode: http.StatusNotImplemented,
			wantContent:    "Unknown metric type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{reader: tt.mockReaderService, writer: tt.mockWriterService}
			cfg := &config.ServerConfig{
				SecretKey: "test key",
			}
			ts := httptest.NewServer(middleware.NewGzipMiddleware()(NewRouter(h, cfg)))
			defer ts.Close()

			bodyBytes, _ := json.Marshal(tt.args.body)

			t.Run("plain request", func(t *testing.T) {
				req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, bytes.NewBuffer(bodyBytes))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept-Encoding", "")

				res, err := ts.Client().Do(req)
				require.NoError(t, err)
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("failed to close res.Body: %v", err)
					}
				}()

				respBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.Equal(t, tt.wantStatusCode, res.StatusCode)
				assert.Equal(t, tt.wantContent+"\n", string(respBody))
			})

			t.Run("gzip response", func(t *testing.T) {
				req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, bytes.NewBuffer(bodyBytes))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept-Encoding", "gzip")

				res, err := ts.Client().Do(req)
				require.NoError(t, err)
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("failed to close res.Body: %v", err)
					}
				}()

				contentEncoding := res.Header.Get("Content-Encoding")
				if contentEncoding != "" && contentEncoding != "gzip" {
					t.Errorf("Expected 'gzip' in Content-Encoding header, but got '%s'", contentEncoding)
				}

				respBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				expectedBody := tt.wantContent + "\n"
				require.NoError(t, err)
				assert.Contains(t, string(respBody), expectedBody)
				assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			})
		})
	}
}

func TestHandler_GetSerializedMetric(t *testing.T) {
	type args struct {
		body   models.Metrics
		method string
		url    string
	}
	tests := []struct {
		name              string
		mockReaderService ReaderServiceInterface
		mockWriterService WriterServiceInterface
		args              args
		wantStatusCode    int
		wantContent       string
	}{
		{
			name: "Test #1 valid counter metric fetch",
			mockReaderService: services.NewService(&mockRepo{
				metrics: map[string]repositories.Metric{
					"counter_metric": {
						Type:  constants.MetricTypeCounter,
						Value: int64(10),
					},
				},
			}),
			mockWriterService: services.NewService(&mockRepo{
				metrics: map[string]repositories.Metric{
					"counter_metric": {
						Type:  constants.MetricTypeCounter,
						Value: int64(10),
					},
				},
			}),
			args: args{
				method: http.MethodPost,
				url:    "/value/",
				body: models.Metrics{
					ID:    "counter_metric",
					MType: constants.MetricTypeCounter,
				},
			},
			wantStatusCode: http.StatusOK,
			wantContent:    `{"id":"counter_metric","type":"counter","delta":10}`,
		},
		{
			name: "Test #2 metric not found",
			mockReaderService: services.NewService(&mockRepo{
				metrics: map[string]repositories.Metric{},
			}),
			mockWriterService: services.NewService(&mockRepo{
				metrics: map[string]repositories.Metric{},
			}),
			args: args{
				method: http.MethodPost,
				url:    "/value/",
				body: models.Metrics{
					ID:    "missing_metric",
					MType: constants.MetricTypeGauge,
				},
			},
			wantStatusCode: http.StatusNotFound,
			wantContent:    "Metric not found",
		},
		{
			name: "Test #3 invalid JSON input",
			mockReaderService: services.NewService(&mockRepo{
				metrics: map[string]repositories.Metric{},
			}),
			mockWriterService: services.NewService(&mockRepo{
				metrics: map[string]repositories.Metric{},
			}),
			args: args{
				method: http.MethodPost,
				url:    "/value/",
				body:   models.Metrics{},
			},
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "Invalid JSON",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{reader: tt.mockReaderService, writer: tt.mockWriterService}
			cfg := &config.ServerConfig{
				SecretKey: "test key",
			}
			ts := httptest.NewServer(middleware.NewGzipMiddleware()(NewRouter(h, cfg)))
			defer ts.Close()

			bodyBytes, _ := json.Marshal(tt.args.body)

			t.Run("plain request", func(t *testing.T) {
				var req *http.Request
				var err error

				if tt.name == "Test #3 invalid JSON input" {
					req, err = http.NewRequest(tt.args.method, ts.URL+tt.args.url, bytes.NewBufferString(`{invalid_json}`))
				} else {
					req, err = http.NewRequest(tt.args.method, ts.URL+tt.args.url, bytes.NewBuffer(bodyBytes))
				}
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept-Encoding", "")

				res, err := ts.Client().Do(req)
				require.NoError(t, err)
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("failed to close res.Body.Close: %v", err)
					}
				}()

				respBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.Equal(t, tt.wantStatusCode, res.StatusCode)
				assert.Equal(t, tt.wantContent+"\n", string(respBody))
			})

			t.Run("gzip response", func(t *testing.T) {
				var req *http.Request
				var err error

				if tt.name == "Test #3 invalid JSON input" {
					req, err = http.NewRequest(tt.args.method, ts.URL+tt.args.url, bytes.NewBufferString(`{invalid_json}`))
				} else {
					req, err = http.NewRequest(tt.args.method, ts.URL+tt.args.url, bytes.NewBuffer(bodyBytes))
				}
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept-Encoding", "gzip")

				res, err := ts.Client().Do(req)
				require.NoError(t, err)
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("failed to close res.Body.Close: %v", err)
					}
				}()

				contentEncoding := res.Header.Get("Content-Encoding")
				if contentEncoding != "gzip" {
					t.Errorf("Expected 'gzip' in Content-Encoding header, but got '%s'", contentEncoding)
				}

				respBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.Equal(t, tt.wantStatusCode, res.StatusCode)
				assert.Contains(t, string(respBody), tt.wantContent+"\n")
			})
		})
	}
}

func TestUpdateSerializedMetrics(t *testing.T) {
	repo := &mockRepo{
		metrics: make(map[string]repositories.Metric),
	}
	service := services.NewService(repo)
	handler := NewHandler(service, service, nil)

	tests := []struct {
		name           string
		metrics        []models.Metrics
		errOnUpdate    bool
		wantStatusCode int
	}{
		{
			name: "Test #1 update metrics batch",
			metrics: []models.Metrics{
				{ID: "test1", MType: constants.MetricTypeGauge, Value: float64Ptr(123.45)},
				{ID: "test2", MType: constants.MetricTypeCounter, Delta: int64Ptr(123)},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Empty metrics array",
			metrics:        []models.Metrics{},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Missing gauge value",
			metrics: []models.Metrics{
				{ID: "test1", MType: constants.MetricTypeGauge},
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Missing counter value",
			metrics: []models.Metrics{
				{ID: "test1", MType: constants.MetricTypeCounter},
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Unknown metric type",
			metrics: []models.Metrics{
				{ID: "test1", MType: "unknown"},
			},
			wantStatusCode: http.StatusNotImplemented,
		},
		{
			name: "Update metrics batch error",
			metrics: []models.Metrics{
				{ID: "test1", MType: constants.MetricTypeGauge, Value: float64Ptr(1.23)},
			},
			errOnUpdate:    true,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.errOnUpdate = tt.errOnUpdate
			recorder := httptest.NewRecorder()
			body, _ := json.Marshal(tt.metrics)
			req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			handler.UpdateSerializedMetrics(recorder, req)
			assert.Equal(t, tt.wantStatusCode, recorder.Code)
		})
	}
}

func BenchmarkUpdateSerializedMetric(b *testing.B) {
	repo := &mockRepo{}
	service := services.NewService(repo)
	h := &Handler{reader: service, writer: service}
	metric := models.Metrics{
		ID:    "test",
		MType: constants.MetricTypeGauge,
		Value: float64Ptr(123.45),
	}
	body, _ := json.Marshal(metric)
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.UpdateSerializedMetric(w, r)
	}
}

func BenchmarkUpdateSerializedMetrics(b *testing.B) {
	repo := &mockRepo{}
	service := services.NewService(repo)
	h := &Handler{reader: service, writer: service}
	metrics := []models.Metrics{
		{ID: "test", MType: constants.MetricTypeGauge, Value: float64Ptr(123.45)},
	}
	body, _ := json.Marshal(metrics)
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.UpdateSerializedMetrics(w, r)
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

type mockRepo struct {
	metrics     map[string]repositories.Metric
	errOnUpdate bool
	errOnGet    bool
}

func (m *mockRepo) SaveMetric(_ context.Context, name string, value interface{}, metricType string) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error")
	}
	if m.metrics == nil {
		m.metrics = make(map[string]repositories.Metric)
	}
	m.metrics[name] = repositories.Metric{
		Type:  metricType,
		Value: value,
	}
	return nil
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

func (m *mockRepo) UpdateMetricsBatch(_ context.Context, metrics map[string]repositories.Metric) error {
	if m.errOnUpdate {
		return fmt.Errorf("mock update error")
	}
	if m.metrics == nil {
		m.metrics = make(map[string]repositories.Metric)
	}
	for name, metric := range metrics {
		m.metrics[name] = metric
	}
	return nil
}

func (m *mockRepo) UpdateGaugeMetric(ctx context.Context, id string, value float64) error {
	return m.SaveMetric(ctx, id, value, constants.MetricTypeGauge)
}

func (m *mockRepo) UpdateCounterMetric(ctx context.Context, id string, delta int64) error {
	return m.SaveMetric(ctx, id, delta, constants.MetricTypeCounter)
}
