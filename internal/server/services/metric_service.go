// Package services provides business logic for working with metrics.
package services

import (
	"context"

	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/utils"
)

// MetricRepository defines the interface for metric storage operations.
type MetricRepository interface {
	SaveMetric(ctx context.Context, metricName string, metricValue interface{}, metricType string) error
	GetMetric(ctx context.Context, metricName string) (repositories.Metric, error)
	GetMetrics(ctx context.Context) (map[string]repositories.Metric, error)
	UpdateMetricsBatch(ctx context.Context, metrics map[string]repositories.Metric) error
}

// Service provides business logic for working with metrics.
type Service struct {
	repo MetricRepository
}

// NewService creates a new Service instance.
func NewService(repo MetricRepository) *Service {
	return &Service{repo: repo}
}

// UpdateGaugeMetric updates a gauge metric.
func (s *Service) UpdateGaugeMetric(ctx context.Context, name string, value float64) error {
	return s.repo.SaveMetric(ctx, name, value, constants.MetricTypeGauge)
}

// UpdateCounterMetric updates a counter metric.
func (s *Service) UpdateCounterMetric(ctx context.Context, name string, value int64) error {
	return s.repo.SaveMetric(ctx, name, value, constants.MetricTypeCounter)
}

// GetMetric retrieves a metric by name.
func (s *Service) GetMetric(ctx context.Context, metricName string) (repositories.Metric, error) {
	return s.repo.GetMetric(ctx, metricName)
}

// GetMetrics retrieves all metrics.
func (s *Service) GetMetrics(ctx context.Context) (map[string]repositories.Metric, error) {
	return s.repo.GetMetrics(ctx)
}

// UpdateMetricsBatch updates a batch of metrics.
func (s *Service) UpdateMetricsBatch(ctx context.Context, metrics map[string]repositories.Metric) error {
	return s.repo.UpdateMetricsBatch(ctx, metrics)
}

// UpdateGaugeMetricWithRetry updates a gauge metric with retry logic.
func (s *Service) UpdateGaugeMetricWithRetry(ctx context.Context, name string, value float64) error {
	return utils.WithRetries(func() error {
		return s.UpdateGaugeMetric(ctx, name, value)
	})
}

// UpdateCounterMetricWithRetry updates a counter metric with retry logic.
func (s *Service) UpdateCounterMetricWithRetry(ctx context.Context, name string, value int64) error {
	return utils.WithRetries(func() error {
		return s.UpdateCounterMetric(ctx, name, value)
	})
}

// GetMetricWithRetry retrieves a metric by name with retry logic.
func (s *Service) GetMetricWithRetry(ctx context.Context, name string) (repositories.Metric, error) {
	var result repositories.Metric
	err := utils.WithRetries(func() error {
		var err error
		result, err = s.GetMetric(ctx, name)
		return err
	})
	return result, err
}

// GetMetricsWithRetry retrieves all metrics with retry logic.
func (s *Service) GetMetricsWithRetry(ctx context.Context) (map[string]repositories.Metric, error) {
	var result map[string]repositories.Metric
	err := utils.WithRetries(func() error {
		var err error
		result, err = s.GetMetrics(ctx)
		return err
	})
	return result, err
}

// UpdateMetricsBatchWithRetry updates a batch of metrics with retry logic.
func (s *Service) UpdateMetricsBatchWithRetry(ctx context.Context, metrics map[string]repositories.Metric) error {
	return utils.WithRetries(func() error {
		return s.UpdateMetricsBatch(ctx, metrics)
	})
}
