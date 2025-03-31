package main

import (
	"context"
	"github.com/a2sh3r/sysmetrics/internal/agent"
	"github.com/a2sh3r/sysmetrics/internal/config"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Printf("Error while creating new config: %v", err)
		return
	}

	cfg.ParseFlags()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	metricAgent := agent.NewAgent(cfg)

	log.Println("Starting agent...")
	metricAgent.Run(ctx)
}
