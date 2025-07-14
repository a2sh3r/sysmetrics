package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/a2sh3r/sysmetrics/internal/agent"
	"github.com/a2sh3r/sysmetrics/internal/config"
)

func main() {
	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Printf("Error while creating new config: %v", err)
		return
	}

	cfg.ParseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metricAgent := agent.NewAgent(cfg)

	log.Println("Starting agent...")

	go metricAgent.Run(ctx)

	<-ctx.Done()
	log.Println("Shutting down agent...")
}
