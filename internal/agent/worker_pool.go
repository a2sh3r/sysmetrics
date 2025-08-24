package agent

import (
	"context"
	"sync"
	"time"

	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"go.uber.org/zap"
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
					logger.Log.Info("Worker stopped due to context cancellation")
					return
				case <-w.ctx.Done():
					logger.Log.Info("Worker stopped due to worker cancellation")
					return
				case m := <-w.metricsChan:
					if err := w.sendFunc(m); err != nil {
						logger.Log.Error("Error sending metrics", zap.Error(err))
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
		logger.Log.Warn("Metrics channel is full, dropping metrics")
	}
}

func (w *MetricsWorker) Stop() {
	logger.Log.Info("Stopping metrics worker...")

	w.cancel()

	close(w.metricsChan)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Log.Info("All workers stopped successfully")
	case <-time.After(5 * time.Second):
		logger.Log.Warn("Worker shutdown timeout reached")
	}
}
