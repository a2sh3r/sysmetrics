package repositories

import "context"

type Storage interface {
	UpdateMetric(ctx context.Context, metricName string, metric Metric) error
	GetMetric(ctx context.Context, metricName string) (Metric, error)
	GetMetrics(ctx context.Context) (map[string]Metric, error)
	UpdateMetricsBatch(ctx context.Context, metrics map[string]Metric) error
}

type Metric struct {
	Type  string
	Value interface{}
}
