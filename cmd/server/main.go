package main

import (
	"log"
	_ "net/http/pprof"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/a2sh3r/sysmetrics/internal/server/startup"
)

func main() {
	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Printf("Error while creating new config: %v", err)
		return
	}
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Printf("Error while creating logger instance: %v", err)
	}

	cfg.ParseFlags()

	if err := startup.RunServer(cfg); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
