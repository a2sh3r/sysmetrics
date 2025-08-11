package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a2sh3r/sysmetrics/internal/agent"
	"github.com/a2sh3r/sysmetrics/internal/config"
)

var buildVersion string
var buildDate string
var buildCommit string

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	log.Printf("Build version: %s", buildVersion)
	log.Printf("Build date: %s", buildDate)
	log.Printf("Build commit: %s", buildCommit)
}

func main() {
	printBuildInfo()
	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Printf("Error while creating new config: %v", err)
		return
	}

	cfg.ParseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	metricAgent := agent.NewAgent(cfg)

	log.Println("Starting agent...")

	go metricAgent.Run(ctx)

	<-ctx.Done()

	log.Println("Shutting down agent gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	select {
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout reached, forcing exit")
	default:
		log.Println("Agent shutdown completed")
	}
}
