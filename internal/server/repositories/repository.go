// Package repositories provides interfaces for metric storage.
package repositories

import "context"

// Storage defines the interface for metric storage backends.
type Storage interface {
	UpdateMetric(ctx context.Context, metricName string, metric Metric) error
	GetMetric(ctx context.Context, metricName string) (Metric, error)
	GetMetrics(ctx context.Context) (map[string]Metric, error)
	UpdateMetricsBatch(ctx context.Context, metrics map[string]Metric) error
}

// Metric represents a single metric with type and value.
type Metric struct {
	Type  string
	Value interface{}
}
