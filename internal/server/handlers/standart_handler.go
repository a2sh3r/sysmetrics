package handlers

import (
	"bytes"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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
		logger.Log.Warn("Failed to validate params", zap.Error(err), zap.String("metricType", metricType), zap.String("metricName", metricName))
		return
	}

	responseMetric, err := h.service.GetMetric(metricName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metric: %s", err), http.StatusNotFound)
		logger.Log.Warn("Metric not found", zap.Error(err), zap.String("metricName", metricName))
		return
	}

	if responseMetric.Type != metricType {
		http.Error(w, fmt.Sprintf("Got metric, but its type differs from the requested one: %s", metricName), http.StatusBadRequest)
		logger.Log.Warn("Metric type mismatch", zap.String("requestedType", metricType), zap.String("actualType", responseMetric.Type), zap.String("metricName", metricName))
		return
	}

	if responseMetric.Value == nil {
		http.Error(w, "Metric value is nil", http.StatusInternalServerError)
		logger.Log.Error("Metric value is nil", zap.String("metricName", metricName))
		return
	}

	metricString, err := formatMetric(nil, responseMetric.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Log.Error("Failed to format metric", zap.Error(err), zap.String("metricName", metricName))
		return
	}

	if _, err := io.WriteString(w, metricString); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write body: %s\n", err), http.StatusInternalServerError)
		logger.Log.Error("Failed to write response body", zap.Error(err), zap.String("metricName", metricName))
		return
	}

	logger.Log.Info("Sending metric value",
		zap.Any("response", responseMetric),
		zap.String("response_body", metricString),
	)
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
		logger.Log.Error("Failed to get metrics", zap.Error(err))
		return
	}

	var metricsBuffer bytes.Buffer
	for metricName, responseMetric := range responseMetrics {
		metricString, err := formatMetric(&metricName, responseMetric.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Log.Error("Failed to format metric", zap.Error(err), zap.String("metricName", metricName))
			return
		}
		metricsBuffer.WriteString(metricString)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := io.Copy(w, &metricsBuffer); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write body: %s", err), http.StatusInternalServerError)
		logger.Log.Error("Failed to write response body", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)

	logger.Log.Info("Sending all metrics",
		zap.Int("metrics_count", len(responseMetrics)),
		zap.String("response_body", metricsBuffer.String()),
	)
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
		logger.Log.Warn("Invalid URL format", zap.String("path", r.URL.Path))
		return
	}

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if err := validateParams(metricType, metricName, metricValue); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusBadRequest)
		logger.Log.Warn("Failed to validate params", zap.Error(err), zap.String("metricType", metricType), zap.String("metricName", metricName), zap.String("metricValue", metricValue))
		return
	}

	switch metricType {
	case constants.MetricTypeGauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusBadRequest)
			logger.Log.Warn("Failed to parse gauge value", zap.Error(err), zap.String("metricValue", metricValue))
			return
		}
		if err := h.service.UpdateGaugeMetric(metricName, value); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusInternalServerError)
			logger.Log.Error("Failed to update gauge metric", zap.Error(err), zap.String("metricName", metricName))
			return
		}
	case constants.MetricTypeCounter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusBadRequest)
			logger.Log.Warn("Failed to parse counter value", zap.Error(err), zap.String("metricValue", metricValue))
			return
		}
		if err := h.service.UpdateCounterMetric(metricName, value); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update metric: %s", err), http.StatusInternalServerError)
			logger.Log.Error("Failed to update counter metric", zap.Error(err), zap.String("metricName", metricName))
			return
		}
	default:
		http.Error(w, fmt.Sprintf("Failed to update metric: invalid metric type: %v", metricType), http.StatusBadRequest)
		logger.Log.Warn("Invalid metric type", zap.String("metricType", metricType))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	msg := fmt.Sprintf("Metric %v is updated successfully with value %v", metricName, metricValue)
	if _, err := io.WriteString(w, msg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write body: %s", err), http.StatusInternalServerError)
		logger.Log.Error("Failed to write response body", zap.Error(err), zap.String("metricName", metricName))
		return
	}

	logger.Log.Info("Metric updated successfully",
		zap.String("metricName", metricName),
		zap.String("metricType", metricType),
		zap.String("metricValue", metricValue),
		zap.String("response_body", msg),
	)
}
