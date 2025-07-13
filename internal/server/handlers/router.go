package handlers

import (
	"net/http"

	"github.com/a2sh3r/sysmetrics/internal/server/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(handler *Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.NewLoggingMiddleware())
	r.Use(middleware.NewGzipMiddleware())

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
	})

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetMetrics)
		r.Post("/update/{metricType}/{metricName}/{metricValue}", handler.UpdateMetric)
		r.Post("/update/", handler.UpdateSerializedMetric)
		r.Get("/value/{metricType}/{metricName}", handler.GetMetric)
		r.Post("/value/", handler.GetSerializedMetric)
		r.Post("/updates/", handler.UpdateSerializedMetrics)
		r.Get("/ping", handler.Ping)
	})

	return r
}
