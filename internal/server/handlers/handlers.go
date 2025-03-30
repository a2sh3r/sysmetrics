package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/server/services/metric"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	service *metric.Service
}

func NewRouter(handler *Handler) chi.Router {
	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
	})

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetMetrics)
		r.Post("/update/{metricType}/{metricName}/{metricValue}", handler.UpdateMetric)
		r.Get("/value/{metricType}/{metricName}", handler.GetMetric)
	})

	return r
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

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if err := validateParams(metricType, metricName, metricValue); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusBadRequest)
	}

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
		http.Error(w, fmt.Sprintf("Failed to update metric: invalid metric type: %v", metricType), http.StatusBadRequest)
		return
	}

	if _, err := io.WriteString(w, fmt.Sprintf("Metric %v is updated successfully with value %v", metricName, metricValue)); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write body: %s", err), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Date", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))

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

	switch v := responseMetric.Value.(type) {
	case int64:
		if _, err := io.WriteString(w, fmt.Sprintf("%d\n", v)); err != nil {
			http.Error(w, fmt.Sprintf("Failed to write body: %s\n", err), http.StatusInternalServerError)
			return
		}
	case float64:
		if _, err := io.WriteString(w, fmt.Sprintf("%g\n", v)); err != nil {
			http.Error(w, fmt.Sprintf("Failed to write body: %s\n", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Unsupported metric value type", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Date", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))

	responseMetrics, err := h.service.GetMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metrics: %s", err), http.StatusInternalServerError)
	}

	var metricsBuffer bytes.Buffer
	for metricName, responseMetric := range responseMetrics {
		switch v := responseMetric.Value.(type) {
		case int64:
			_, err := fmt.Fprintf(&metricsBuffer, "%s %d\n", metricName, v)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to write body: %s", err), http.StatusInternalServerError)
				return
			}
		case float64:
			_, err := fmt.Fprintf(&metricsBuffer, "%s %g\n", metricName, v)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to write body: %s", err), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Unsupported metric value type", http.StatusInternalServerError)
			return
		}
	}

	if _, err := io.Copy(w, &metricsBuffer); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write body: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func validateParams(params ...string) error {
	for _, p := range params {
		if p == "" {
			return errors.New("one of metric parameters are null")
		}
	}
	return nil
}
