package handlers

import (
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/server/services/metric"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	service *metric.Service
}

func NewHandler(service *metric.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Date", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
		return
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateGaugeMetric(metricName, value); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusInternalServerError)
			return
		}
	case "counter":
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
		http.Error(w, "Failed to update metric: invalid metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Metric %v of %v type has written successfully", metricName, metricType)
}
