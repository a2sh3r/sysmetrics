package agent

import (
	"context"
	"log"
	"sync"

	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
)

type MetricsWorker struct {
	rateLimit   int64
	metricsChan chan *metrics.Metrics
	wg          sync.WaitGroup
	sendFunc    func(*metrics.Metrics) error
}

func NewMetricsWorker(rateLimit int64, sendFunc func(*metrics.Metrics) error) *MetricsWorker {
	return &MetricsWorker{
		metricsChan: make(chan *metrics.Metrics, rateLimit*2),
		rateLimit:   rateLimit,
		sendFunc:    sendFunc,
	}
}

func (w *MetricsWorker) Start(ctx context.Context) {
	w.wg.Add(int(w.rateLimit))

	for i := int64(0); i < w.rateLimit; i++ {
		go func() {
			defer w.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case m := <-w.metricsChan:
					if err := w.sendFunc(m); err != nil {
						log.Printf("Error sending metrics: %v", err)
						continue
					}
				}
			}
		}()
	}
}

func (w *MetricsWorker) SendMetrics(metrics *metrics.Metrics) {
	select {
	case w.metricsChan <- metrics:
	default:
	}
}

func (w *MetricsWorker) Stop() {
	w.wg.Wait()
	close(w.metricsChan)
}
