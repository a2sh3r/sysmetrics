// Package repositories provides repository implementations for metrics storage.
package repositories

import "context"

// MetricRepo implements the MetricRepository interface using a Storage backend.
type MetricRepo struct {
	storage Storage
}

// NewMetricRepo creates a new MetricRepo instance.
func NewMetricRepo(storage Storage) *MetricRepo {
	return &MetricRepo{storage: storage}
}

// SaveMetric saves a metric to the storage.
func (r *MetricRepo) SaveMetric(ctx context.Context, metricName string, metricValue interface{}, metricType string) error {
	return r.storage.UpdateMetric(ctx, metricName, Metric{Type: metricType, Value: metricValue})
}

// GetMetric retrieves a metric from the storage.
func (r *MetricRepo) GetMetric(ctx context.Context, metricName string) (Metric, error) {
	return r.storage.GetMetric(ctx, metricName)
}

// GetMetrics retrieves all metrics from the storage.
func (r *MetricRepo) GetMetrics(ctx context.Context) (map[string]Metric, error) {
	return r.storage.GetMetrics(ctx)
}

// UpdateMetricsBatch updates a batch of metrics in the storage.
func (r *MetricRepo) UpdateMetricsBatch(ctx context.Context, metrics map[string]Metric) error {
	return r.storage.UpdateMetricsBatch(ctx, metrics)
}
