package handler

import (
	"errors"
	"frontdev333/bookshelf/internal/domain"
	"frontdev333/bookshelf/internal/service"
	"log/slog"
	"net/http"
	"strings"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest

	if err := decode(r, &req); err != nil {
		slog.Error("Register() decode body", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "Internal error")
		return
	}

	errDetails := make([]domain.ErrorDetail, 0)

	if strings.TrimSpace(req.Username) == "" {
		errDetails = append(errDetails, domain.ErrorDetail{
			Field:   "Username",
			Message: "Username must not be empty",
		})
	}

	if strings.TrimSpace(req.Password) == "" {
		errDetails = append(errDetails, domain.ErrorDetail{
			Field:   "Password",
			Message: "Password must not be empty",
		})
	}

	if strings.TrimSpace(req.Email) == "" {
		errDetails = append(errDetails, domain.ErrorDetail{
			Field:   "Email",
			Message: "Email must not be empty",
		})
	}

	if len(errDetails) > 0 {
		writeValidationError(w, r, errDetails)
		return
	}
	errDetails = []domain.ErrorDetail{}

	resp, err := h.services.User.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}
		if errors.Is(err, service.ErrUserExists) {
			errDetails = append(errDetails, domain.ErrorDetail{
				Field:   "Email",
				Message: "User with this email already exists",
			})
			writeValidationError(w, r, errDetails)
			return
		}
		if errors.Is(err, service.ErrInvalidPassword) {
			errDetails = append(errDetails, domain.ErrorDetail{
				Field:   "Password",
				Message: "Password must be at least 8 characters",
			})
			writeValidationError(w, r, errDetails)
			return
		}
		if errors.Is(err, service.ErrInvalidUsername) {
			errDetails = append(errDetails, domain.ErrorDetail{
				Field:   "Username",
				Message: "Username must be at least 3 characters",
			})
			writeValidationError(w, r, errDetails)
			return
		}
		if errors.Is(err, service.ErrInvalidEmail) {
			errDetails = append(errDetails, domain.ErrorDetail{
				Field:   "Email",
				Message: "Invalid email format",
			})
			writeValidationError(w, r, errDetails)
			return
		}
		if errors.Is(err, service.ErrUsernameExists) {
			errDetails = append(errDetails, domain.ErrorDetail{
				Field:   "Username",
				Message: "Username already exists",
			})
			writeValidationError(w, r, errDetails)
			return
		}
		slog.Error("Register() service error", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "Internal error")
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq domain.LoginRequest
	if err := decode(r, &loginReq); err != nil {
		slog.Error("Login() decode", "error", err)
		return
	}

	errDetails := make([]domain.ErrorDetail, 0)

	if strings.TrimSpace(loginReq.Email) == "" {
		errDetails = append(errDetails, domain.ErrorDetail{
			Field:   "Username",
			Message: "Username must be at least 3 characters",
		})
	}

	if strings.TrimSpace(loginReq.Password) == "" {
		errDetails = append(errDetails, domain.ErrorDetail{
			Field:   "Password",
			Message: "Password must not be empty",
		})
	}

	if len(errDetails) > 0 {
		writeValidationError(w, r, errDetails)
		return
	}
	errDetails = []domain.ErrorDetail{}

	resp, err := h.services.User.Login(r.Context(), loginReq)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}
		if errors.Is(err, service.ErrInvalidEmail) {
			errDetails = append(errDetails, domain.ErrorDetail{
				Field:   "Email",
				Message: "Invalid email format",
			})
			writeValidationError(w, r, errDetails)
			return
		}
		slog.Error("Login() service error", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "Internal error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	u, err := h.services.User.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "", "Server error. Try later.")
		slog.Error("GetCurrentUser()", "error", err)
		return
	}

	writeJSON(w, http.StatusOK, u.ToPublic())
	return
}

func (h *Handler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	var req domain.UpdateUserRequest

	if err := decode(r, &req); err != nil {
		slog.Error("UpdateCurrentUser() decode", "error", err)
		return
	}

	u, err := h.services.User.Update(r.Context(), userID, req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "", "Server error. Try later.")
		slog.Error("UpdateCurrentUser()", "error", err)
		return
	}

	writeJSON(w, http.StatusOK, u.ToPublic())

}
