package main

import (
	"context"
	"github.com/a2sh3r/sysmetrics/internal/agent"
	"github.com/a2sh3r/sysmetrics/internal/agent/config"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.NewConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	metricAgent := agent.NewAgent(cfg)

	log.Println("Starting agent...")
	metricAgent.Run(ctx)
}
