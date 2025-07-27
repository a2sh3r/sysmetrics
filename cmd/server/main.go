package main

import (
	"log"
	_ "net/http/pprof"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/a2sh3r/sysmetrics/internal/server/startup"
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
	println("Build version:", buildVersion)
	println("Build date:", buildDate)
	println("Build commit:", buildCommit)
}

func main() {
	printBuildInfo()
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
