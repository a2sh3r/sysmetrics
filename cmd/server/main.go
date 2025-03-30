package main

import (
	"github.com/a2sh3r/sysmetrics/cmd/server/flag"
	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/server/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services/metric"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	"net/http"
)

func main() {
	cfg := config.NewServerConfig()

	flag.ParseFlags(cfg)

	memStorage := memstorage.NewMemStorage()

	metricRepo := repositories.NewMetricRepo(memStorage)

	metricService := metric.NewService(metricRepo)

	handler := handlers.NewHandler(metricService)

	//log.Printf("Server is staring on address: %s", cfg.Address)

	if err := http.ListenAndServe(cfg.Address, handlers.NewRouter(handler)); err != nil {
		panic(err)
	}
}
