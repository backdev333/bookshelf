package handler

import (
	"bookshelf/auth-service/internal/domain"
	"context"
	"encoding/json"
	"frontdev333/bookshelf/internal/service"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const userIDKey contextKey = "userID"
const requestID contextKey = "requestID"
const version = "1.0.0"

type Handler struct {
	services  *service.Service
	jwtSecret string
}

type pong struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

func New(services *service.Service, jwtSecret string) *Handler {
	return &Handler{
		services:  services,
		jwtSecret: jwtSecret,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	v := pong{
		Status:    "ok",
		Version:   version,
		Timestamp: time.Now(),
	}

	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	type readyPong struct {
		pong
		Database string `json:"database"`
	}

	v := readyPong{
		pong: pong{Status: "ok",
			Version:   version,
			Timestamp: time.Now(),
		},
		Database: "ok",
	}

	writeJSON(w, http.StatusOK, v)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	bytes, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		slog.Error("writeJSON", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(bytes); err != nil {
		slog.Error("writeJSON", "error", err)
		return
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	reqID := middleware.GetReqID(r.Context())

	errResp := domain.ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: reqID,
	}

	writeJSON(w, status, errResp)
}

func writeValidationError(w http.ResponseWriter, r *http.Request, details []domain.ErrorDetail) {
	reqID := middleware.GetReqID(r.Context())
	errResp := domain.ErrorResponse{
		Code:      "VALIDATION ERROR",
		Message:   "Invalid Input",
		Details:   details,
		RequestID: reqID,
	}

	writeJSON(w, http.StatusUnprocessableEntity, errResp)
}

func decode(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}

func getUserID(ctx context.Context) string {
	return ctx.Value(userIDKey).(string)
}
