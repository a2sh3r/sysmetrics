package handlers

import (
	"encoding/json"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/a2sh3r/sysmetrics/internal/models"
	"go.uber.org/zap"
	"net/http"
)

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
		if err := h.writer.UpdateGaugeMetric(m.ID, *m.Value); err != nil {
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
		if err := h.writer.UpdateCounterMetric(m.ID, *m.Delta); err != nil {
			logger.Log.Error("Failed to update counter", zap.String("metric_id", m.ID), zap.Error(err))
			http.Error(w, "Failed to update counter", http.StatusInternalServerError)
			return
		}
	default:
		logger.Log.Warn("Unsupported metric type", zap.String("type", m.MType))
		http.Error(w, "Unknown metric type", http.StatusNotImplemented)
		return
	}

	updated, err := h.reader.GetMetric(m.ID)
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

	stored, err := h.reader.GetMetric(m.ID)
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
