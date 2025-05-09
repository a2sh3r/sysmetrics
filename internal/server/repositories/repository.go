package repositories

type Storage interface {
	UpdateMetric(metricName string, metric Metric) error
	GetMetric(metricName string) (Metric, error)
	GetMetrics() (map[string]Metric, error)
}

type Metric struct {
	Type  string
	Value interface{}
}
