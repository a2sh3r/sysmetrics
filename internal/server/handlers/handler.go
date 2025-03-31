package handlers

import (
	"github.com/a2sh3r/sysmetrics/internal/server/services"
)

type Handler struct {
	service *services.Service
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{service: service}
}
