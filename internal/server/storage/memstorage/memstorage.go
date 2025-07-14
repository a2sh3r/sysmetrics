// Package memstorage provides an in-memory implementation of the Storage interface.
package memstorage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

var (
	ErrMetricNotFound    = errors.New("metric not found")
	ErrStorageNil        = errors.New("MemStorage is nil")
	ErrMetricsMapNil     = errors.New("metrics map is nil")
	ErrMetricInvalidType = errors.New("invalid value type for metric")
	ErrMetricInvalidName = errors.New("invalid metric error")
)

// MemStorage implements in-memory storage for metrics.
type MemStorage struct {
	metrics map[string]repositories.Metric
	mu      sync.RWMutex
}

// NewMemStorage creates a new MemStorage instance.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]repositories.Metric),
	}
}

// GetMetric retrieves a metric from memory storage.
func (ms *MemStorage) GetMetric(ctx context.Context, metricName string) (repositories.Metric, error) {
	if ctx.Err() != nil {
		return repositories.Metric{}, ctx.Err()
	}

	if ms == nil {
		return repositories.Metric{}, ErrStorageNil
	}

	if ms.metrics == nil {
		return repositories.Metric{}, ErrMetricsMapNil
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if m, ok := ms.metrics[metricName]; ok {
		return m, nil
	}
	return repositories.Metric{}, ErrMetricNotFound
}

// GetMetrics retrieves all metrics from memory storage.
func (ms *MemStorage) GetMetrics(ctx context.Context) (map[string]repositories.Metric, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if ms == nil {
		return map[string]repositories.Metric{}, ErrStorageNil
	}

	if ms.metrics == nil {
		return map[string]repositories.Metric{}, ErrMetricsMapNil
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.metrics, nil
}

func (ms *MemStorage) UpdateMetric(ctx context.Context, metricName string, metric repositories.Metric) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if metricName == "" {
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

	if metric.Type != constants.MetricTypeCounter && metric.Type != constants.MetricTypeGauge {
		return ErrMetricInvalidType
	}

	existingMetric, exists := ms.metrics[metricName]

	if !exists {
		ms.metrics[metricName] = metric
		return nil
	}

	if existingMetric.Type != metric.Type {
		return ErrMetricInvalidType
	}

	switch metric.Type {
	case constants.MetricTypeCounter:
		err := ms.updateCounterMetric(&existingMetric, metric)
		if err != nil {
			return err
		}
	case constants.MetricTypeGauge:
		err := ms.updateGaugeMetric(&existingMetric, metric)
		if err != nil {
			return err
		}
	default:
		return ErrMetricInvalidType
	}
	ms.metrics[metricName] = existingMetric
	return nil
}

func (ms *MemStorage) updateCounterMetric(existingMetric *repositories.Metric, newMetric repositories.Metric) error {
	if newMetric.Type != constants.MetricTypeCounter {
		return ErrMetricInvalidType
	}
	newValue, ok := newMetric.Value.(int64)
	if !ok {
		return ErrMetricInvalidType
	}

	existingValue, ok := existingMetric.Value.(int64)
	if !ok {
		existingValue = 0
	}

	existingMetric.Value = existingValue + newValue
	return nil
}

func (ms *MemStorage) updateGaugeMetric(existingMetric *repositories.Metric, newMetric repositories.Metric) error {
	if newMetric.Type != constants.MetricTypeGauge {
		return ErrMetricInvalidType
	}
	newValue, ok := newMetric.Value.(float64)
	if !ok {
		return fmt.Errorf("6, %T, %v", newMetric.Value, ok)
	}

	existingMetric.Value = newValue
	return nil
}

func (ms *MemStorage) UpdateMetricsBatch(ctx context.Context, metrics map[string]repositories.Metric) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if ms == nil {
		return fmt.Errorf("storage is nil")
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.metrics == nil {
		return ErrMetricsMapNil
	}

	for name, metric := range metrics {
		if name == "" {
			return ErrMetricInvalidName
		}

		if metric.Type != constants.MetricTypeCounter && metric.Type != constants.MetricTypeGauge {
			return ErrMetricInvalidType
		}

		existingMetric, exists := ms.metrics[name]

		if !exists {
			ms.metrics[name] = metric
			continue
		}

		if existingMetric.Type != metric.Type {
			return ErrMetricInvalidType
		}

		switch metric.Type {
		case constants.MetricTypeCounter:
			newValue, ok := metric.Value.(int64)
			if !ok {
				return ErrMetricInvalidType
			}

			existingValue, ok := existingMetric.Value.(int64)
			if !ok {
				existingValue = 0
			}

			existingMetric.Value = existingValue + newValue
		case constants.MetricTypeGauge:
			newValue, ok := metric.Value.(float64)
			if !ok {
				return fmt.Errorf("invalid gauge value type: %T", metric.Value)
			}
			existingMetric.Value = newValue
		default:
			return ErrMetricInvalidType
		}
		ms.metrics[name] = existingMetric
	}

	return nil
}
