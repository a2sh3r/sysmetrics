package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
)

type MetricsWorker struct {
	rateLimit   int64
	metricsChan chan *metrics.Metrics
	wg          sync.WaitGroup
	sendFunc    func(*metrics.Metrics) error
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewMetricsWorker(rateLimit int64, sendFunc func(*metrics.Metrics) error) *MetricsWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &MetricsWorker{
		metricsChan: make(chan *metrics.Metrics, rateLimit*2),
		rateLimit:   rateLimit,
		sendFunc:    sendFunc,
		ctx:         ctx,
		cancel:      cancel,
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
					log.Println("Worker stopped due to context cancellation")
					return
				case <-w.ctx.Done():
					log.Println("Worker stopped due to worker cancellation")
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
		log.Println("Metrics channel is full, dropping metrics")
	}
}

func (w *MetricsWorker) Stop() {
	log.Println("Stopping metrics worker...")

	w.cancel()

	close(w.metricsChan)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers stopped successfully")
	case <-time.After(5 * time.Second):
		log.Println("Worker shutdown timeout reached")
	}
}
