package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/models"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
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
	tests := []struct {
		name           string
		mockService    ServiceInterface
		args           args
		wantStatusCode int
		wantContent    string
	}{
		{
			name: "Test #1 valid gauge metric update",
			mockService: services.NewService(&mockRepo{
				metrics: make(map[string]repositories.Metric),
			}),
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
			name: "Test #2 missing gauge value",
			mockService: services.NewService(&mockRepo{
				metrics: make(map[string]repositories.Metric),
			}),
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
			name: "Test #3 unsupported metric type",
			mockService: services.NewService(&mockRepo{
				metrics: make(map[string]repositories.Metric),
			}),
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
			h := &Handler{
				service: tt.mockService,
			}
			ts := httptest.NewServer(NewRouter(h))
			defer ts.Close()

			bodyBytes, _ := json.Marshal(tt.args.body)
			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, bytes.NewBuffer(bodyBytes))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			respBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			assert.Equal(t, tt.wantContent+"\n", string(respBody))
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
		name           string
		mockService    ServiceInterface
		args           args
		wantStatusCode int
		wantContent    string
	}{
		{
			name: "Test #1 valid counter metric fetch",
			mockService: services.NewService(&mockRepo{
				metrics: map[string]repositories.Metric{
					"counter_metric": {
						Type:  constants.MetricTypeCounter,
						Value: int64(10),
					},
				},
			}),
			args: args{
				method: http.MethodGet,
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
			mockService: services.NewService(&mockRepo{
				metrics: map[string]repositories.Metric{},
			}),
			args: args{
				method: http.MethodGet,
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
			mockService: services.NewService(&mockRepo{
				metrics: map[string]repositories.Metric{},
			}),
			args: args{
				method: http.MethodGet,
				url:    "/value/",
				body:   models.Metrics{},
			},
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "Invalid JSON",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.mockService,
			}
			ts := httptest.NewServer(NewRouter(h))
			defer ts.Close()

			var req *http.Request
			var err error

			if tt.name == "Test #3 invalid JSON input" {
				req, err = http.NewRequest(tt.args.method, ts.URL+tt.args.url, bytes.NewBufferString(`{invalid_json}`))
			} else {
				bodyBytes, _ := json.Marshal(tt.args.body)
				req, err = http.NewRequest(tt.args.method, ts.URL+tt.args.url, bytes.NewBuffer(bodyBytes))
			}
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			respBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			assert.Equal(t, tt.wantContent+"\n", string(respBody))
		})
	}
}
