package main

import (
	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/a2sh3r/sysmetrics/internal/server/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Printf("Error while creating new config: %v", err)
		return
	}

	cfg.ParseFlags()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Logger init failed: %v", err)
	}

	memStorage := memstorage.NewMemStorage()

	metricRepo := repositories.NewMetricRepo(memStorage)

	metricService := services.NewService(metricRepo)

	handler := handlers.NewHandler(metricService)

	logger.Log.Info("Server is staring",
		zap.String("address", cfg.Address),
	)

	if err := http.ListenAndServe(cfg.Address, handlers.NewRouter(handler)); err != nil {
		panic(err)
	}
}
