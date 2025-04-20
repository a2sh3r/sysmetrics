package handlers

import (
	"bytes"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	setHeaders(w)

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	if err := validateParams(metricType, metricName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusBadRequest)
	}

	responseMetric, err := h.service.GetMetric(metricName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metric: %s", err), http.StatusNotFound)
		return
	}

	if responseMetric.Type != metricType {
		http.Error(w, fmt.Sprintf("Got metric, but its type differs from the requested one: %s", metricName), http.StatusBadRequest)
		return
	}

	if responseMetric.Value == nil {
		http.Error(w, "Metric value is nil", http.StatusInternalServerError)
		return
	}

	metricString, err := formatMetric(nil, responseMetric.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := io.WriteString(w, metricString); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write body: %s\n", err), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	setHeaders(w)

	responseMetrics, err := h.service.GetMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metrics: %s", err), http.StatusInternalServerError)
	}

	var metricsBuffer bytes.Buffer
	for metricName, responseMetric := range responseMetrics {
		metricString, err := formatMetric(&metricName, responseMetric.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		metricsBuffer.WriteString(metricString)
	}

	if _, err := io.Copy(w, &metricsBuffer); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write body: %s", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

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
