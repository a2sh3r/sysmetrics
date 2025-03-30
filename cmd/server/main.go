package main

import (
	"github.com/a2sh3r/sysmetrics/internal/server/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services/metric"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	"log"
	"net/http"
)

func main() {
	log.Println("Server is staring")

	memStorage := memstorage.NewMemStorage()

	metricRepo := repositories.NewMetricRepo(memStorage)

	metricService := metric.NewService(metricRepo)

	handler := handlers.NewHandler(metricService)

	if err := http.ListenAndServe(":8080", handlers.NewRouter(handler)); err != nil {
		panic(err)
	}
}
