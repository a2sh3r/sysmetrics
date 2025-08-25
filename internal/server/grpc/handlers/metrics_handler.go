// Package handlers provides gRPC request handlers for metrics operations.
package handlers

import (
	"context"

	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	pb "github.com/a2sh3r/sysmetrics/pkg/api/grpc/metrics"
)

// MetricsHandler implements the gRPC MetricsService
type MetricsHandler struct {
	pb.UnimplementedMetricsServiceServer
	service *services.Service
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(service *services.Service) *MetricsHandler {
	return &MetricsHandler{service: service}
}

// UpdateMetric updates a single metric
func (h *MetricsHandler) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	metric := req.GetMetric()

	var err error
	if metric.GetType() == constants.MetricTypeGauge {
		err = h.service.UpdateGaugeMetric(ctx, metric.GetId(), metric.GetValue())
	} else if metric.GetType() == constants.MetricTypeCounter {
		err = h.service.UpdateCounterMetric(ctx, metric.GetId(), metric.GetDelta())
	} else {
		return &pb.UpdateMetricResponse{
			Success: false,
			Error:   "invalid metric type",
		}, nil
	}

	if err != nil {
		return &pb.UpdateMetricResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.UpdateMetricResponse{Success: true}, nil
}

// UpdateMetricsBatch updates multiple metrics
func (h *MetricsHandler) UpdateMetricsBatch(ctx context.Context, req *pb.UpdateMetricsBatchRequest) (*pb.UpdateMetricsBatchResponse, error) {
	metrics := req.GetMetrics()

	for _, metric := range metrics {
		var err error
		if metric.GetType() == constants.MetricTypeGauge {
			err = h.service.UpdateGaugeMetric(ctx, metric.GetId(), metric.GetValue())
		} else if metric.GetType() == constants.MetricTypeCounter {
			err = h.service.UpdateCounterMetric(ctx, metric.GetId(), metric.GetDelta())
		} else {
			return &pb.UpdateMetricsBatchResponse{
				Success: false,
				Error:   "invalid metric type",
			}, nil
		}

		if err != nil {
			return &pb.UpdateMetricsBatchResponse{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
	}

	return &pb.UpdateMetricsBatchResponse{Success: true}, nil
}

// GetMetric retrieves a single metric
func (h *MetricsHandler) GetMetric(ctx context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	metric, err := h.service.GetMetric(ctx, req.GetId())
	if err != nil {
		return &pb.GetMetricResponse{
			Found: false,
			Error: err.Error(),
		}, nil
	}

	var value float64
	var delta int64

	switch metric.Type {
	case constants.MetricTypeGauge:
		if v, ok := metric.Value.(float64); ok {
			value = v
		}
	case constants.MetricTypeCounter:
		if v, ok := metric.Value.(int64); ok {
			delta = v
		}
	}

	pbMetric := &pb.Metric{
		Id:    req.GetId(),
		Type:  metric.Type,
		Value: value,
		Delta: delta,
	}

	return &pb.GetMetricResponse{
		Metric: pbMetric,
		Found:  true,
	}, nil
}

// GetMetrics retrieves all metrics
func (h *MetricsHandler) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	metrics, err := h.service.GetMetrics(ctx)
	if err != nil {
		return &pb.GetMetricsResponse{Error: err.Error()}, nil
	}

	var pbMetrics []*pb.Metric
	for name, metric := range metrics {
		var value float64
		var delta int64

		switch metric.Type {
		case constants.MetricTypeGauge:
			if v, ok := metric.Value.(float64); ok {
				value = v
			}
		case constants.MetricTypeCounter:
			if v, ok := metric.Value.(int64); ok {
				delta = v
			}
		}

		pbMetric := &pb.Metric{
			Id:    name,
			Type:  metric.Type,
			Value: value,
			Delta: delta,
		}
		pbMetrics = append(pbMetrics, pbMetric)
	}

	return &pb.GetMetricsResponse{Metrics: pbMetrics}, nil
}
