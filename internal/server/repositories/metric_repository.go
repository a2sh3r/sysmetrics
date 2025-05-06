package repositories

import "context"

type MetricRepo struct {
	storage Storage
}

func NewMetricRepo(storage Storage) *MetricRepo {
	return &MetricRepo{storage: storage}
}

func (r *MetricRepo) SaveMetric(ctx context.Context, metricName string, metricValue interface{}, metricType string) error {
	return r.storage.UpdateMetric(ctx, metricName, Metric{Type: metricType, Value: metricValue})
}

func (r *MetricRepo) GetMetric(ctx context.Context, metricName string) (Metric, error) {
	return r.storage.GetMetric(ctx, metricName)
}

func (r *MetricRepo) GetMetrics(ctx context.Context) (map[string]Metric, error) {
	return r.storage.GetMetrics(ctx)
}

func (r *MetricRepo) UpdateMetricsBatch(ctx context.Context, metrics map[string]Metric) error {
	return r.storage.UpdateMetricsBatch(ctx, metrics)
}
