package metric

import (
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
)

type Service struct {
	storage *memstorage.MemStorage
}

func NewService(storage *memstorage.MemStorage) *Service {
	return &Service{storage: storage}
}

func (s *Service) UpdateGaugeMetric(name string, value float64) error {
	return s.storage.UpdateMetric(name, memstorage.Metric{Type: "gauge", Value: value})
}

func (s *Service) UpdateCounterMetric(name string, value int64) error {
	return s.storage.UpdateMetric(name, memstorage.Metric{Type: "counter", Value: float64(value)})
}
