package main

import (
	"context"
	"github.com/a2sh3r/sysmetrics/cmd/agent/flag"
	"github.com/a2sh3r/sysmetrics/internal/agent"
	"github.com/a2sh3r/sysmetrics/internal/config"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.NewAgentConfig()

	flag.ParseFlags(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	metricAgent := agent.NewAgent(cfg)

	log.Println("Starting agent...")
	metricAgent.Run(ctx)
}
