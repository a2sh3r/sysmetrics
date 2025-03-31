package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
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
