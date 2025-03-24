package main

import (
	"github.com/a2sh3r/sysmetrics/internal/server/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services/metric"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	"net/http"
)

func main() {
	memStorage := memstorage.NewMemStorage()

	metricRepo := repositories.NewMetricRepo(memStorage)

	metricService := metric.NewService(metricRepo)

	handler := handlers.NewHandler(metricService)

	http.HandleFunc("/update/", handler.UpdateMetric)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
