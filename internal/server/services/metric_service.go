package services

import (
	"context"

	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/utils"
)

type MetricRepository interface {
	SaveMetric(ctx context.Context, metricName string, metricValue interface{}, metricType string) error
	GetMetric(ctx context.Context, metricName string) (repositories.Metric, error)
	GetMetrics(ctx context.Context) (map[string]repositories.Metric, error)
	UpdateMetricsBatch(ctx context.Context, metrics map[string]repositories.Metric) error
}

type Service struct {
	repo MetricRepository
}

func NewService(repo MetricRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdateGaugeMetric(ctx context.Context, name string, value float64) error {
	return s.repo.SaveMetric(ctx, name, value, constants.MetricTypeGauge)
}

func (s *Service) UpdateCounterMetric(ctx context.Context, name string, value int64) error {
	return s.repo.SaveMetric(ctx, name, value, constants.MetricTypeCounter)
}

func (s *Service) GetMetric(ctx context.Context, metricName string) (repositories.Metric, error) {
	return s.repo.GetMetric(ctx, metricName)
}

func (s *Service) GetMetrics(ctx context.Context) (map[string]repositories.Metric, error) {
	return s.repo.GetMetrics(ctx)
}

func (s *Service) UpdateMetricsBatch(ctx context.Context, metrics map[string]repositories.Metric) error {
	return s.repo.UpdateMetricsBatch(ctx, metrics)
}

func (s *Service) UpdateGaugeMetricWithRetry(ctx context.Context, name string, value float64) error {
	return utils.WithRetries(func() error {
		return s.UpdateGaugeMetric(ctx, name, value)
	})
}

func (s *Service) UpdateCounterMetricWithRetry(ctx context.Context, name string, value int64) error {
	return utils.WithRetries(func() error {
		return s.UpdateCounterMetric(ctx, name, value)
	})
}

func (s *Service) GetMetricWithRetry(ctx context.Context, name string) (repositories.Metric, error) {
	var result repositories.Metric
	err := utils.WithRetries(func() error {
		var err error
		result, err = s.GetMetric(ctx, name)
		return err
	})
	return result, err
}

func (s *Service) GetMetricsWithRetry(ctx context.Context) (map[string]repositories.Metric, error) {
	var result map[string]repositories.Metric
	err := utils.WithRetries(func() error {
		var err error
		result, err = s.GetMetrics(ctx)
		return err
	})
	return result, err
}

func (s *Service) UpdateMetricsBatchWithRetry(ctx context.Context, metrics map[string]repositories.Metric) error {
	return utils.WithRetries(func() error {
		return s.UpdateMetricsBatch(ctx, metrics)
	})
}
