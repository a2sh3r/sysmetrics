package main

import (
	"github.com/a2sh3r/sysmetrics/internal/server/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/storage"
	"net/http"
)

func main() {
	memStorage := storage.NewMemStorage()

	handler := handlers.NewHandler(memStorage)

	http.HandleFunc("/update/", handler.UpdateMetric)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
