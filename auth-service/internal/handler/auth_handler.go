package handler

import (
	domain2 "bookshelf/auth-service/internal/domain"
	"errors"
	handler2 "frontdev333/bookshelf/internal/handler"
	"frontdev333/bookshelf/internal/service"
	"log/slog"
	"net/http"
	"strings"
)

func (h *handler2.Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain2.RegisterRequest

	if err := handler2.decode(r, &req); err != nil {
		slog.Error("Register() decode body", "error", err)
		handler2.writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	errDetails := make([]domain2.ErrorDetail, 0)

	if strings.TrimSpace(req.Username) == "" {
		errDetails = append(errDetails, domain2.ErrorDetail{
			Field:   "Username",
			Message: "Username must not be empty",
		})
	}

	if strings.TrimSpace(req.Password) == "" {
		errDetails = append(errDetails, domain2.ErrorDetail{
			Field:   "Password",
			Message: "Password must not be empty",
		})
	}

	if strings.TrimSpace(req.Email) == "" {
		errDetails = append(errDetails, domain2.ErrorDetail{
			Field:   "Email",
			Message: "Email must not be empty",
		})
	}

	if len(errDetails) > 0 {
		handler2.writeValidationError(w, r, errDetails)
		return
	}
	errDetails = []domain2.ErrorDetail{}

	resp, err := h.services.User.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			handler2.writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}
		if errors.Is(err, service.ErrUserExists) {
			errDetails = append(errDetails, domain2.ErrorDetail{
				Field:   "Email",
				Message: "User with this email already exists",
			})
			handler2.writeValidationError(w, r, errDetails)
			return
		}
		if errors.Is(err, service.ErrInvalidPassword) {
			errDetails = append(errDetails, domain2.ErrorDetail{
				Field:   "Password",
				Message: "Password must be at least 8 characters",
			})
			handler2.writeValidationError(w, r, errDetails)
			return
		}
		if errors.Is(err, service.ErrInvalidUsername) {
			errDetails = append(errDetails, domain2.ErrorDetail{
				Field:   "Username",
				Message: "Username must be at least 3 characters",
			})
			handler2.writeValidationError(w, r, errDetails)
			return
		}
		if errors.Is(err, service.ErrInvalidEmail) {
			errDetails = append(errDetails, domain2.ErrorDetail{
				Field:   "Email",
				Message: "Invalid email format",
			})
			handler2.writeValidationError(w, r, errDetails)
			return
		}
		if errors.Is(err, service.ErrUsernameExists) {
			errDetails = append(errDetails, domain2.ErrorDetail{
				Field:   "Username",
				Message: "Username already exists",
			})
			handler2.writeValidationError(w, r, errDetails)
			return
		}
		slog.Error("Register() service error", "error", err)
		handler2.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal error")
		return
	}

	handler2.writeJSON(w, http.StatusCreated, resp)
}

func (h *handler2.Handler) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq domain2.LoginRequest
	if err := handler2.decode(r, &loginReq); err != nil {
		slog.Error("Login() decode", "error", err)
		handler2.writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	errDetails := make([]domain2.ErrorDetail, 0)

	if strings.TrimSpace(loginReq.Email) == "" {
		errDetails = append(errDetails, domain2.ErrorDetail{
			Field:   "Username",
			Message: "Username must be at least 3 characters",
		})
	}

	if strings.TrimSpace(loginReq.Password) == "" {
		errDetails = append(errDetails, domain2.ErrorDetail{
			Field:   "Password",
			Message: "Password must not be empty",
		})
	}

	if len(errDetails) > 0 {
		handler2.writeValidationError(w, r, errDetails)
		return
	}
	errDetails = []domain2.ErrorDetail{}

	resp, err := h.services.User.Login(r.Context(), loginReq)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			handler2.writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}
		if errors.Is(err, service.ErrInvalidEmail) {
			errDetails = append(errDetails, domain2.ErrorDetail{
				Field:   "Email",
				Message: "Invalid email format",
			})
			handler2.writeValidationError(w, r, errDetails)
			return
		}
		slog.Error("Login() service error", "error", err)
		handler2.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal error")
		return
	}

	handler2.writeJSON(w, http.StatusOK, resp)
}

func (h *handler2.Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := handler2.getUserID(r.Context())

	u, err := h.services.User.GetByID(r.Context(), userID)
	if err != nil {
		handler2.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Server error. Try later.")
		slog.Error("GetCurrentUser()", "error", err)
		return
	}

	handler2.writeJSON(w, http.StatusOK, u.ToPublic())
	return
}

func (h *handler2.Handler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := handler2.getUserID(r.Context())
	var req domain2.UpdateUserRequest

	if err := handler2.decode(r, &req); err != nil {
		slog.Error("UpdateCurrentUser() decode", "error", err)
		handler2.writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	u, err := h.services.User.Update(r.Context(), userID, req)
	if err != nil {
		handler2.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Server error. Try later.")
		slog.Error("UpdateCurrentUser()", "error", err)
		return
	}

	handler2.writeJSON(w, http.StatusOK, u.ToPublic())

}
