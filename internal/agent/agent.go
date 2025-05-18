package agent

import (
	"context"
	"github.com/a2sh3r/sysmetrics/internal/agent/collector"
	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/a2sh3r/sysmetrics/internal/agent/sender"
	"github.com/a2sh3r/sysmetrics/internal/config"
	"log"
	"time"
)

type Agent struct {
	collector      *collector.Collector
	sender         *sender.Sender
	pollInterval   time.Duration
	reportInterval time.Duration
}

func NewAgent(cfg *config.AgentConfig) *Agent {
	return &Agent{
		collector:      collector.NewCollector(),
		sender:         sender.NewSender(cfg.Address, cfg.SecretKey),
		pollInterval:   time.Second * time.Duration(cfg.PollInterval),
		reportInterval: time.Second * time.Duration(cfg.ReportInterval),
	}
}

func (a *Agent) Run(ctx context.Context) {
	metricsChan := make(chan *metrics.Metrics, int(a.reportInterval/a.pollInterval)+1) // Размер буфера зависит от интервалов сбора и отправки метрик. В случае если интервал будет меньше 1, то буфер равен 1.

	go func() {
		ticker := time.NewTicker(a.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metricsChan <- a.collector.CollectMetrics()
			case <-ctx.Done():
				close(metricsChan)
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(a.reportInterval)
		defer ticker.Stop()

		var metricsBatch []*metrics.Metrics

		for {
			select {
			case <-ticker.C:
				if len(metricsBatch) > 0 {
					if err := a.sender.SendMetricsWithRetries(ctx, metricsBatch); err != nil {
						log.Printf("Error sending metrics: %v", err)
					}
					metricsBatch = nil
				}
			case metric := <-metricsChan:
				metricsBatch = append(metricsBatch, metric)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	log.Println("Agent stopped")
}
