package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/a2sh3r/sysmetrics/internal/agent/sender"
	"github.com/a2sh3r/sysmetrics/internal/config"
)

type Agent struct {
	cfg     *config.AgentConfig
	metrics *metrics.Metrics
	worker  *MetricsWorker
	sender  *sender.Sender
	mu      sync.RWMutex
}

func NewAgent(cfg *config.AgentConfig) *Agent {
	return &Agent{
		cfg:     cfg,
		metrics: metrics.NewMetrics(),
		sender:  sender.NewSender(cfg.Address, cfg.SecretKey),
	}
}

func (a *Agent) Run(ctx context.Context) {
	a.worker = NewMetricsWorker(a.cfg.RateLimit, a.sendMetrics)
	a.worker.Start(ctx)

	go func() {
		ticker := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.mu.Lock()
				a.metrics = metrics.NewMetrics()
				a.mu.Unlock()
				a.worker.SendMetrics(a.metrics)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.mu.Lock()
				if err := a.metrics.UpdateSystemMetrics(); err != nil {
					log.Printf("Error updating system metrics: %v", err)
				}
				a.mu.Unlock()
				a.worker.SendMetrics(a.metrics)
			}
		}
	}()

	<-ctx.Done()
	a.worker.Stop()
}

func (a *Agent) sendMetrics(m *metrics.Metrics) error {
	return a.sender.SendMetricsWithRetries(context.Background(), []*metrics.Metrics{m})
}
