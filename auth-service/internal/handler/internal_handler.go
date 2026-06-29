package handler

import (
	"bookshelf/auth-svc/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type InternalHandler struct {
	svc *service.UserService
}

type VerifyResponse struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Error     string    `json:"error"`
}

func NewInternalHandler(svc *service.UserService) *InternalHandler {
	return &InternalHandler{
		svc: svc,
	}
}

func (h *InternalHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var token struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		slog.Error("InternalHandler.VerifyToekn() json.NewDecoder()", "error", err)
		writeJSON(
			w,
			http.StatusBadRequest,
			VerifyResponse{
				Valid: false,
				Error: err.Error(),
			},
		)
		return
	}

	claims, err := h.svc.ValidateToken(token.Token)
	if err != nil {
		slog.Error("InternalHandler.VerifyToekn() h.UService.ValidateToken", "error", err)
		writeJSON(
			w,
			http.StatusUnauthorized,
			VerifyResponse{
				Valid: false,
				Error: err.Error(),
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		VerifyResponse{
			Valid:     true,
			UserID:    claims.Subject,
			ExpiresAt: claims.ExpiresAt.Time,
			Error:     "",
		},
	)
	return
}
