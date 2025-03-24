package repositories

type MetricRepo struct {
	storage Storage
}

func NewMetricRepo(storage Storage) *MetricRepo {
	return &MetricRepo{storage: storage}
}

func (r *MetricRepo) SaveMetric(name string, value interface{}, metricType string) error {
	return r.storage.UpdateMetric(name, Metric{Type: metricType, Value: value})
}

func (r *MetricRepo) GetMetric(name string) (Metric, error) {
	return r.storage.GetMetric(name)
}
