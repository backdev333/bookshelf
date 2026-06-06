package handler

import (
	"bookshelf/auth-svc/internal/domain"
	"bookshelf/auth-svc/internal/service"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const version = "1.0.0"

type AuthHandler struct {
	svc       *service.UserService
	jwtSecret string
}

func New(service *service.UserService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		svc:       service,
		jwtSecret: jwtSecret,
	}
}

type pong struct {
	Status    string    `json:"status,omitempty"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
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

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	reqID := middleware.GetReqID(r.Context())

	errResp := domain.ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: reqID,
	}

	msg := struct {
		Error domain.ErrorResponse `json:"error"`
	}{
		Error: errResp,
	}

	writeJSON(w, status, msg)
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
