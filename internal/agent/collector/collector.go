package collector

import "github.com/a2sh3r/sysmetrics/internal/agent/metrics"

type Collector struct {
	pollCount int64
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) CollectMetrics() *metrics.Metrics {
	collectedMetrics := metrics.NewMetrics()
	c.pollCount++
	collectedMetrics.PollCount = c.pollCount
	return collectedMetrics
}
