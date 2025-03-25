package main

import (
	"github.com/a2sh3r/sysmetrics/internal/agent/collector"
	"github.com/a2sh3r/sysmetrics/internal/agent/config"
	"github.com/a2sh3r/sysmetrics/internal/agent/sender"
	"log"
	"time"
)

func main() {
	cfg := config.NewConfig()

	metricCollector := collector.NewCollector()
	metricSender := sender.NewSender(cfg.ServerAddress)

	for {
		metrics := metricCollector.CollectMetrics()

		if err := metricSender.SendMetrics(metrics); err != nil {
			log.Printf("Error sending metrics: %v\n", err)
		}

		time.Sleep(cfg.ReportInterval)
	}
}
