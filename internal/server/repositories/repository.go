package repositories

type Storage interface {
	UpdateMetric(name string, metric Metric) error
	GetMetric(name string) (Metric, error)
}

type MetricRepository interface {
	SaveMetric(name string, value interface{}, metricType string) error
	GetMetric(name string) (Metric, error)
}

type Metric struct {
	Type  string
	Value interface{}
}
