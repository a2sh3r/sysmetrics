package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/a2sh3r/sysmetrics/internal/models"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

// UpdateSerializedMetric handles POST requests to update a metric using a JSON body.
func (h *Handler) UpdateSerializedMetric(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log.Warn("Failed to decode JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	logger.Log.Info("Received metric update request", zap.Any("metric", m))

	switch m.MType {
	case constants.MetricTypeGauge:
		if m.Value == nil {
			logger.Log.Warn("Missing value for gauge", zap.String("metric_id", m.ID))
			http.Error(w, "Missing gauge value", http.StatusBadRequest)
			return
		}
		if err := h.writer.UpdateGaugeMetricWithRetry(r.Context(), m.ID, *m.Value); err != nil {
			logger.Log.Error("Failed to update gauge", zap.String("metric_id", m.ID), zap.Error(err))
			http.Error(w, "Failed to update gauge", http.StatusInternalServerError)
			return
		}
	case constants.MetricTypeCounter:
		if m.Delta == nil {
			logger.Log.Warn("Missing delta for counter", zap.String("metric_id", m.ID))
			http.Error(w, "Missing counter delta", http.StatusBadRequest)
			return
		}
		if err := h.writer.UpdateCounterMetricWithRetry(r.Context(), m.ID, *m.Delta); err != nil {
			logger.Log.Error("Failed to update counter", zap.String("metric_id", m.ID), zap.Error(err))
			http.Error(w, "Failed to update counter", http.StatusInternalServerError)
			return
		}
	default:
		logger.Log.Warn("Unsupported metric type", zap.String("type", m.MType))
		http.Error(w, "Unknown metric type", http.StatusNotImplemented)
		return
	}

	updated, err := h.reader.GetMetricWithRetry(r.Context(), m.ID)
	if err != nil {
		logger.Log.Error("Metric not found after update", zap.String("metric_id", m.ID), zap.Error(err))
		http.Error(w, "Metric not found after update", http.StatusNotFound)
		return
	}

	response := convertMetricToModel(m.ID, updated)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

// GetSerializedMetric handles POST requests to get a metric value using a JSON body.
func (h *Handler) GetSerializedMetric(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log.Warn("Failed to decode JSON for value request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.ID == "" || m.MType == "" {
		logger.Log.Warn("Missing fields in value request", zap.Any("metric", m))
		http.Error(w, "Missing metric ID or type", http.StatusBadRequest)
		return
	}

	stored, err := h.reader.GetMetricWithRetry(r.Context(), m.ID)
	if err != nil {
		logger.Log.Warn("Metric not found", zap.String("metric_id", m.ID), zap.Error(err))
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	response := convertMetricToModel(m.ID, stored)

	logger.Log.Info("Sending metric value", zap.Any("response", response))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

// UpdateSerializedMetrics handles POST requests to update multiple metrics using a JSON array.
func (h *Handler) UpdateSerializedMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics []models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		logger.Log.Warn("Failed to decode JSON for batch update", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	repoMetrics := make(map[string]repositories.Metric)
	for _, m := range metrics {
		switch m.MType {
		case constants.MetricTypeGauge:
			if m.Value == nil {
				logger.Log.Warn("Missing value for gauge", zap.String("metric_id", m.ID))
				http.Error(w, "Missing gauge value", http.StatusBadRequest)
				return
			}
			repoMetrics[m.ID] = repositories.Metric{
				Type:  constants.MetricTypeGauge,
				Value: *m.Value,
			}
		case constants.MetricTypeCounter:
			if m.Delta == nil {
				logger.Log.Warn("Missing delta for counter", zap.String("metric_id", m.ID))
				http.Error(w, "Missing counter delta", http.StatusBadRequest)
				return
			}
			if existing, ok := repoMetrics[m.ID]; ok && existing.Type == constants.MetricTypeCounter {
				repoMetrics[m.ID] = repositories.Metric{
					Type:  constants.MetricTypeCounter,
					Value: existing.Value.(int64) + *m.Delta,
				}
			} else {
				repoMetrics[m.ID] = repositories.Metric{
					Type:  constants.MetricTypeCounter,
					Value: *m.Delta,
				}
			}
		default:
			logger.Log.Warn("Unsupported metric type", zap.String("type", m.MType))
			http.Error(w, "Unknown metric type", http.StatusNotImplemented)
			return
		}
	}

	if err := h.writer.UpdateMetricsBatchWithRetry(r.Context(), repoMetrics); err != nil {
		logger.Log.Error("Failed to update metrics batch", zap.Error(err))
		http.Error(w, "Failed to update metrics", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
