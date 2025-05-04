package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/models"
	"github.com/a2sh3r/sysmetrics/internal/server/middleware"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
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
					Value: floatPointer(123.12),
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
			ts := httptest.NewServer(middleware.NewGzipMiddleware()(NewRouter(h)))
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
			ts := httptest.NewServer(middleware.NewGzipMiddleware()(NewRouter(h)))
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
