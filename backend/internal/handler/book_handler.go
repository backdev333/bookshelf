package handler

import (
	"errors"
	"frontdev333/bookshelf/internal/domain"
	"frontdev333/bookshelf/internal/service"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListBook(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")

	f := domain.BookFilter{
		Search: &search,
		Sort:   &sort,
		Order:  &order,
		Page:   &page,
		Limit:  &limit,
	}

	list, err := h.services.Book.List(r.Context(), f)
	if err != nil {
		slog.Error("book_handler.ListBook()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	bookId := chi.URLParam(r, "bookId")

	b, err := h.services.Book.GetByID(r.Context(), bookId)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "BOOK_NOT_FOUND", "book not found")
			return
		}
		slog.Error("book_handler.GetBook()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	var req domain.CreateBookRequest

	if err := decode(r, &req); err != nil {
		slog.Error("book_handler.CreateBook()", "error", err)
		writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	errDetails := make([]domain.ErrorDetail, 0)

	if strings.TrimSpace(req.Title) == "" {
		errDetails = append(errDetails, domain.ErrorDetail{
			Field:   "Title",
			Message: "Title is required",
		})
	}

	if strings.TrimSpace(req.Author) == "" {
		errDetails = append(errDetails, domain.ErrorDetail{
			Field:   "Author",
			Message: "Author is required",
		})
	}

	if len(errDetails) > 0 {
		writeValidationError(w, r, errDetails)
		return
	}
	errDetails = []domain.ErrorDetail{}

	b, err := h.services.Book.Create(r.Context(), userID, req)
	if err != nil {
		slog.Error("book_handler.CreateBook()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}

	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	bookID := chi.URLParam(r, "bookId")

	var req domain.UpdateBookRequest

	if err := decode(r, &req); err != nil {
		slog.Error("book_handler.UpdateBook()", "error", err)
		writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	b, err := h.services.Book.Update(r.Context(), userID, bookID, req)
	if err != nil {
		if errors.Is(err, service.ErrNotBookOwner) {
			writeError(w, r, http.StatusForbidden, "NOT_BOOK_OWNER", "you are not the book owner")
			return
		}

		if errors.Is(err, service.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "BOOK_NOT_FOUND", "book not found")
			return
		}

		slog.Error("book_handler.UpdateBook()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}

	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	bookID := chi.URLParam(r, "bookId")

	if err := h.services.Book.Delete(r.Context(), userID, bookID); err != nil {
		if errors.Is(err, service.ErrNotBookOwner) {
			writeError(w, r, http.StatusForbidden, "NOT_BOOK_OWNER", "you are not the book owner")
			return
		}

		if errors.Is(err, service.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "BOOK_NOT_FOUND", "book not found")
			return
		}

		slog.Error("book_handler.DeleteBook()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
