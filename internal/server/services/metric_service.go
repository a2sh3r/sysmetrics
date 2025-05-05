package services

import (
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/utils"
)

type MetricRepository interface {
	SaveMetric(metricName string, metricValue interface{}, metricType string) error
	GetMetric(metricName string) (repositories.Metric, error)
	GetMetrics() (map[string]repositories.Metric, error)
	UpdateMetricsBatch(metrics map[string]repositories.Metric) error
}

type Service struct {
	repo MetricRepository
}

func NewService(repo MetricRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdateGaugeMetric(name string, value float64) error {
	return s.repo.SaveMetric(name, value, constants.MetricTypeGauge)
}

func (s *Service) UpdateCounterMetric(name string, value int64) error {
	return s.repo.SaveMetric(name, value, constants.MetricTypeCounter)
}

func (s *Service) GetMetric(metricName string) (repositories.Metric, error) {
	return s.repo.GetMetric(metricName)
}

func (s *Service) GetMetrics() (map[string]repositories.Metric, error) {
	return s.repo.GetMetrics()
}

func (s *Service) UpdateMetricsBatch(metrics map[string]repositories.Metric) error {
	return s.repo.UpdateMetricsBatch(metrics)
}

func (s *Service) UpdateGaugeMetricWithRetry(name string, value float64) error {
	return utils.WithRetries(func() error {
		return s.UpdateGaugeMetric(name, value)
	})
}

func (s *Service) UpdateCounterMetricWithRetry(name string, value int64) error {
	return utils.WithRetries(func() error {
		return s.UpdateCounterMetric(name, value)
	})
}

func (s *Service) GetMetricWithRetry(name string) (repositories.Metric, error) {
	var result repositories.Metric
	err := utils.WithRetries(func() error {
		var err error
		result, err = s.GetMetric(name)
		return err
	})
	return result, err
}

func (s *Service) GetMetricsWithRetry() (map[string]repositories.Metric, error) {
	var result map[string]repositories.Metric
	err := utils.WithRetries(func() error {
		var err error
		result, err = s.GetMetrics()
		return err
	})
	return result, err
}

func (s *Service) UpdateMetricsBatchWithRetry(metrics map[string]repositories.Metric) error {
	return utils.WithRetries(func() error {
		return s.UpdateMetricsBatch(metrics)
	})
}
