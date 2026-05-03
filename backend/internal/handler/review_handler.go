package handler

import (
	"errors"
	"frontdev333/bookshelf/internal/domain"
	"frontdev333/bookshelf/internal/service"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListBookReviews(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		slog.Error("review_handler.ListBookReviews()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "unknown error")
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		slog.Error("review_handler.ListBookReviews()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "unknown error")
		return
	}

	reviewList, err := h.services.Review.ListByBookID(r.Context(), bookID, page, limit)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "", "book not found")
			return
		}
		slog.Error("review_handler.ListBookReviews()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "unknown error")
		return
	}

	writeJSON(w, http.StatusOK, reviewList)
}

func (h *Handler) GetReview(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "reviewId")
	resp, err := h.services.Review.GetByID(r.Context(), reviewID)
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			writeError(w, r, http.StatusNotFound, "", "review not found")
			return
		}
		slog.Error("review_handler.GetReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "unknown error")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	userID := getUserID(r.Context())

	var req domain.CreateReviewRequest

	if err := decode(r, &req); err != nil {
		slog.Error("review_handler.CreateReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "unknown error")
		return
	}

	if req.Rating > 5 || req.Rating < 1 {
		writeError(w, r, http.StatusBadRequest, "", "rating must be between 1 and 5")
		return
	}

	if len(req.Content) < 10 {
		writeError(w, r, http.StatusBadRequest, "", "content must be at least 10 characters")
		return
	}

	resp, err := h.services.Review.Create(r.Context(), userID, bookID, req)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "", "book not found")
			return
		}

		if errors.Is(err, service.ErrAlreadyReviewed) {
			writeError(w, r, http.StatusConflict, "", "you have already reviewed this book")
			return
		}

		slog.Error("review_handler.CreateReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "unknown error")
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	reviewID := chi.URLParam(r, "reviewId")

	var req domain.UpdateReviewRequest

	if err := decode(r, &req); err != nil {
		slog.Error("review_handler.UpdateReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "unknown error")
		return
	}

	resp, err := h.services.Review.Update(r.Context(), userID, reviewID, req)
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			writeError(w, r, http.StatusNotFound, "", "review not found")
			return
		}

		if errors.Is(err, service.ErrNotReviewOwner) {
			writeError(w, r, http.StatusForbidden, "", "you are not the owner of this review")
			return
		}

		if errors.Is(err, service.ErrInvalidRating) {
			writeError(w, r, http.StatusBadRequest, "", "rating must be between 1 and 5")
			return
		}

		if errors.Is(err, service.ErrReviewContentTooShort) {
			writeError(w, r, http.StatusBadRequest, "", "content must be at least 10 characters")
			return
		}

		slog.Error("review_handler.UpdateReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "unknown error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	reviewID := chi.URLParam(r, "reviewId")

	if err := h.services.Review.Delete(r.Context(), userID, reviewID); err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			writeError(w, r, http.StatusNotFound, "", "review not found")
			return
		}

		if errors.Is(err, service.ErrNotReviewOwner) {
			writeError(w, r, http.StatusForbidden, "", "you are not the owner of this review")
			return
		}

		slog.Error("review_handler.DeleteReview()", "error", err)
		writeError(w, r, http.StatusInternalServerError, "", "unknown error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
