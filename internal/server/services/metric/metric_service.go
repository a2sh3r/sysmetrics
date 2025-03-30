package metric

import (
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

type Service struct {
	repo repositories.MetricRepository
}

func NewService(repo repositories.MetricRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdateGaugeMetric(name string, value float64) error {
	return s.repo.SaveMetric(name, value, "gauge")
}

func (s *Service) UpdateCounterMetric(name string, value int64) error {
	return s.repo.SaveMetric(name, value, "counter")
}

func (s *Service) GetMetric(metricName string) (repositories.Metric, error) {
	return s.repo.GetMetric(metricName)
}

func (s *Service) GetMetrics() (map[string]repositories.Metric, error) {
	return s.repo.GetMetrics()
}
