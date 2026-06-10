package handler

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/service"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(service *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{service}
}

func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "book_id")

	var err error
	var page int
	var limit int

	getPage := r.URL.Query().Get("page")
	if getPage != "" {
		page, err = strconv.Atoi(getPage)
		if err != nil {
			slog.Error("review_handler.ListBookReviews()", "error", err, "page", page)
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
			return
		}
	}

	if page <= 0 {
		page = 1
	}

	getLimit := r.URL.Query().Get("limit")

	if getLimit != "" {
		limit, err = strconv.Atoi(getLimit)
		if err != nil {
			slog.Error("review_handler.ListBookReviews()", "error", err, "limit", limit)
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
			return
		}
	}

	if limit <= 0 {
		limit = 10
	}

	reviewList, err := h.svc.ListByBookID(r.Context(), bookID, page, limit)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "BOOK_NOT_FOUND", "book not found")
			return
		}
		slog.Error("review_handler.ListBookReviews()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}

	writeJSON(w, http.StatusOK, reviewList)
}

func (h *ReviewHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "reviewId")
	resp, err := h.svc.GetByID(r.Context(), reviewID)
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			writeError(w, r, http.StatusNotFound, "REVIEW_NOT_FOUND", "review not found")
			return
		}
		slog.Error("review_handler.GetReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	userID := getUserID(r.Context())

	var req domain.CreateReviewRequest

	if err := decode(r, &req); err != nil {
		slog.Error("review_handler.CreateReview()", "error", err)
		writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if req.Rating > 5 || req.Rating < 1 {
		writeError(w, r, http.StatusBadRequest, "INVALID_RATING", "rating must be between 1 and 5")
		return
	}

	if len(req.Content) < 10 {
		writeError(w, r, http.StatusBadRequest, "REVIEW_CONTENT_TOO_SHORT", "content must be at least 10 characters")
		return
	}

	resp, err := h.svc.Create(r.Context(), userID, bookID, req)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "BOOK_NOT_FOUND", "book not found")
			return
		}

		if errors.Is(err, service.ErrAlreadyReviewed) {
			writeError(w, r, http.StatusConflict, "ALREADY_REVIEWED", "you have already reviewed this book")
			return
		}

		slog.Error("review_handler.CreateReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	reviewID := chi.URLParam(r, "reviewId")

	var req domain.UpdateReviewRequest

	if err := decode(r, &req); err != nil {
		slog.Error("review_handler.UpdateReview()", "error", err)
		writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	resp, err := h.svc.Update(r.Context(), userID, reviewID, req)
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			writeError(w, r, http.StatusNotFound, "REVIEW_NOT_FOUND", "review not found")
			return
		}

		if errors.Is(err, service.ErrNotReviewOwner) {
			writeError(w, r, http.StatusForbidden, "NOT_REVIEW_OWNER", "you are not the owner of this review")
			return
		}

		if errors.Is(err, service.ErrInvalidRating) {
			writeError(w, r, http.StatusBadRequest, "INVALID_RATING", "rating must be between 1 and 5")
			return
		}

		if errors.Is(err, service.ErrReviewContentTooShort) {
			writeError(w, r, http.StatusBadRequest, "REVIEW_CONTENT_TOO_SHORT", "content must be at least 10 characters")
			return
		}

		slog.Error("review_handler.UpdateReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	reviewID := chi.URLParam(r, "reviewId")

	if err := h.svc.Delete(r.Context(), userID, reviewID); err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			writeError(w, r, http.StatusNotFound, "REVIEW_NOT_FOUND", "review not found")
			return
		}

		if errors.Is(err, service.ErrNotReviewOwner) {
			writeError(w, r, http.StatusForbidden, "NOT_REVIEW_OWNER", "you are not the owner of this review")
			return
		}

		slog.Error("review_handler.DeleteReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
