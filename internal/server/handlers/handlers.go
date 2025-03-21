package handlers

import (
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/server/storage"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	storage *storage.MemStorage
}

func NewHandler(storage *storage.MemStorage) *Handler {
	return &Handler{storage: storage}
}

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
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

	var metric storage.Metric

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge metric value", http.StatusBadRequest)
			return
		}
		metric = storage.Metric{Type: "gauge", Value: value}
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid gauge metric value", http.StatusBadRequest)
			return
		}
		metric = storage.Metric{Type: "counter", Value: value}
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if err := h.storage.UpdateMetric(metricName, metric); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Metric %v of %v type has written successfully", metricName, metricType)
}
