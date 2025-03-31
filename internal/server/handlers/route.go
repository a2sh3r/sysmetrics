package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func NewRouter(handler *Handler) chi.Router {
	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
	})

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetMetrics)
		r.Post("/update/{metricType}/{metricName}/{metricValue}", handler.UpdateMetric)
		r.Get("/value/{metricType}/{metricName}", handler.GetMetric)
	})

	return r
}
