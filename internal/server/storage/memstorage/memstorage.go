package memstorage

import (
	"errors"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"sync"
)

var (
	ErrMetricNotFound    = errors.New("metric not found")
	ErrStorageNil        = errors.New("MemStorage is nil")
	ErrMetricsMapNil     = errors.New("metrics map is nil")
	ErrMetricInvalidType = errors.New("invalid value type for metric")
	ErrMetricInvalidName = errors.New("invalid metric error")
)

type MemStorage struct {
	metrics map[string]repositories.Metric
	mu      sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]repositories.Metric),
	}
}

func (ms *MemStorage) GetMetric(name string) (repositories.Metric, error) {
	if ms == nil {
		return repositories.Metric{}, ErrStorageNil
	}

	if ms.metrics == nil {
		return repositories.Metric{}, ErrMetricsMapNil
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if m, ok := ms.metrics[name]; ok {
		return m, nil
	}
	return repositories.Metric{}, ErrMetricNotFound
}

func (ms *MemStorage) UpdateMetric(name string, metric repositories.Metric) error {
	if name == "" {
		return ErrMetricInvalidName
	}
	if ms == nil {
		return ErrStorageNil
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.metrics == nil {
		return ErrMetricsMapNil
	}

	if metric.Type != "counter" && metric.Type != "gauge" {
		return ErrMetricInvalidType
	}

	existingMetric, exists := ms.metrics[name]

	if !exists {
		ms.metrics[name] = metric
		return nil
	}

	if existingMetric.Type != metric.Type {
		return ErrMetricInvalidType
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
		return ErrMetricInvalidType
	}
	ms.metrics[name] = existingMetric
	return nil
}

func (ms *MemStorage) updateCounterMetric(existingMetric *repositories.Metric, newMetric repositories.Metric) error {
	if newMetric.Type != "counter" {
		return ErrMetricInvalidType
	}
	newValue, ok := newMetric.Value.(int64)
	if !ok {
		return ErrMetricInvalidType
	}

	if existingMetric.Value == nil {
		existingMetric.Value = newValue
		return nil
	}

	existingValue, ok := existingMetric.Value.(int64)
	if !ok {
		return ErrMetricInvalidType
	}

	existingMetric.Value = existingValue + newValue
	return nil
}

func (ms *MemStorage) updateGaugeMetric(existingMetric *repositories.Metric, newMetric repositories.Metric) error {
	if newMetric.Type != "gauge" {
		return ErrMetricInvalidType
	}
	newValue, ok := newMetric.Value.(float64)
	if !ok {
		return fmt.Errorf("6, %T, %v", newMetric.Value, ok)
	}

	existingMetric.Value = newValue
	return nil
}
