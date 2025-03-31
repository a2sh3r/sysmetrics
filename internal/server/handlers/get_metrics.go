package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

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
