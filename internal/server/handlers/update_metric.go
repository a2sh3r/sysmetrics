package handlers

import (
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	setHeaders(w)

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
		return
	}

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if err := validateParams(metricType, metricName, metricValue); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusBadRequest)
	}

	switch metricType {
	case constants.MetricTypeGauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateGaugeMetric(metricName, value); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusInternalServerError)
			return
		}
	case constants.MetricTypeCounter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateCounterMetric(metricName, value); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("Failed to update metric: invalid metric type: %v", metricType), http.StatusBadRequest)
		return
	}

	if _, err := io.WriteString(w, fmt.Sprintf("Metric %v is updated successfully with value %v", metricName, metricValue)); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write body: %s", err), http.StatusInternalServerError)
		return
	}
}
