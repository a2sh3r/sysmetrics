package main

import (
	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/server/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
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

	memStorage := memstorage.NewMemStorage()

	metricRepo := repositories.NewMetricRepo(memStorage)

	metricService := services.NewService(metricRepo)

	handler := handlers.NewHandler(metricService)

	log.Printf("Server is staring on address: %s", cfg.Address)

	if err := http.ListenAndServe(cfg.Address, handlers.NewRouter(handler)); err != nil {
		panic(err)
	}
}
