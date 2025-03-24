package handlers

import (
	"github.com/a2sh3r/sysmetrics/internal/server/services/metric"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_UpdateMetric(t *testing.T) {
	storage := memstorage.NewMemStorage()
	metricService := metric.NewService(storage)
	handler := NewHandler(metricService)

	tests := []struct {
		name       string
		url        string
		statusCode int
	}{
		{
			name:       "valid gauge metric",
			url:        "/update/gauge/test_gauge_metric/10.5",
			statusCode: http.StatusOK,
		},
		{
			name:       "valid counter metric",
			url:        "/update/counter/test_counter_metric/10",
			statusCode: http.StatusOK,
		},
		{
			name:       "invalid metric type",
			url:        "/update/unknown/test_metric/10",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid gauge metric value",
			url:        "/update/gauge/test_metric/abc",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid counter metric value",
			url:        "/update/counter/test_metric/abc",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid URL format",
			url:        "/update",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			handler.UpdateMetric(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}
