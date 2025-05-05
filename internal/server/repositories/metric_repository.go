package repositories

type MetricRepo struct {
	storage Storage
}

func NewMetricRepo(storage Storage) *MetricRepo {
	return &MetricRepo{storage: storage}
}

func (r *MetricRepo) SaveMetric(metricName string, metricValue interface{}, metricType string) error {
	return r.storage.UpdateMetric(metricName, Metric{Type: metricType, Value: metricValue})
}

func (r *MetricRepo) GetMetric(metricName string) (Metric, error) {
	return r.storage.GetMetric(metricName)
}

func (r *MetricRepo) GetMetrics() (map[string]Metric, error) {
	return r.storage.GetMetrics()
}

func (r *MetricRepo) UpdateMetricsBatch(metrics map[string]Metric) error {
	return r.storage.UpdateMetricsBatch(metrics)
}
