package handler

import (
	"bookshelf/auth-service/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

const version = "1.0.0"

type Handler struct {
	service   *service.UserService
	jwtSecret string
}

func New(service *service.UserService, jwtSecret string) *Handler {
	return &Handler{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

type pong struct {
	Status    string    `json:"status,omitempty"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	v := pong{
		Status:    "ok",
		Version:   version,
		Timestamp: time.Now(),
	}

	writeJSON(w, http.StatusOK, v)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		slog.Error("writeJSON", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(dataBytes); err != nil {
		slog.Error("writeJSON", "error", err)
		return
	}
}
