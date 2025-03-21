package storage

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrMetricNotFound    = errors.New("metric not found")
	ErrStorageNil        = errors.New("MemStorage is nil")
	ErrMetricsMapNil     = errors.New("metrics map is nil")
	ErrMetricTypeInvalid = errors.New("invalid value type for metric")
)

type MemStorage struct {
	metrics map[string]Metric
	mu      sync.RWMutex
}

type Metric struct {
	Type  string
	Value interface{}
}

type MetricInterface interface {
	AddMetric(value interface{}) error
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]Metric),
	}
}

func (ms *MemStorage) GetMetric(name string) (Metric, error) {
	if ms == nil {
		return Metric{}, ErrStorageNil
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.metrics == nil {
		return Metric{}, ErrMetricsMapNil
	}

	if m, ok := ms.metrics[name]; ok {
		return m, nil
	}
	return Metric{}, ErrMetricNotFound
}

func (ms *MemStorage) UpdateMetric(name string, metric Metric) error {
	if ms == nil {
		return ErrStorageNil
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.metrics == nil {
		return ErrMetricsMapNil
	}

	if metric.Type != "counter" && metric.Type != "gauge" {
		return ErrMetricTypeInvalid
	}

	existingMetric, exists := ms.metrics[name]

	if !exists {
		ms.metrics[name] = metric
		return nil
	}

	if existingMetric.Type != metric.Type {
		return ErrMetricTypeInvalid
	}

	switch metric.Type {
	case "counter":
		err := ms.updateCounterMetric(&existingMetric, metric)
		if err != nil {
			return err
		}
	case "gauge":
		err := ms.updateGaugeMetric(&existingMetric, metric)
		if err != nil {
			return err
		}
	default:
		return ErrMetricTypeInvalid
	}
	ms.metrics[name] = existingMetric
	return nil
}

func (ms *MemStorage) updateCounterMetric(existingMetric *Metric, newMetric Metric) error {
	newValue, ok := newMetric.Value.(int64)
	if !ok {
		return ErrMetricTypeInvalid
	}

	if existingMetric.Value == nil {
		existingMetric.Value = newValue
		return nil
	}

	existingValue, ok := existingMetric.Value.(int64)
	if !ok {
		return ErrMetricTypeInvalid
	}

	existingMetric.Value = existingValue + newValue
	return nil
}

func (ms *MemStorage) updateGaugeMetric(existingMetric *Metric, newMetric Metric) error {
	newValue, ok := newMetric.Value.(float64)
	if !ok {
		return fmt.Errorf("6, %T, %v", newMetric.Value, ok)
	}

	existingMetric.Value = newValue
	return nil
}
