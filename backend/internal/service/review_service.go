package service

import (
	"context"
	"database/sql"
	"errors"
	"frontdev333/bookshelf/internal/domain"
	"frontdev333/bookshelf/internal/repository"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrReviewNotFound        = errors.New("review not found")
	ErrNotReviewOwner        = errors.New("not review owner")
	ErrAlreadyReviewed       = errors.New("already reviewed")
	ErrInvalidRating         = errors.New("rating must be between 1 and 5")
	ErrReviewContentTooShort = errors.New("review content must be at least 10 characters")
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	bookRepo   *repository.BookRepository
	userRepo   *repository.UserRepository
}

func (s *ReviewService) Create(ctx context.Context, userID, bookID string, req domain.CreateReviewRequest) (*domain.ReviewResponse, error) {
	if _, err := s.bookRepo.GetByID(ctx, bookID); err != nil {
		return nil, ErrBookNotFound
	}

	reviewed, err := s.reviewRepo.UserHasReviewedBook(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}

	if reviewed {
		return nil, ErrAlreadyReviewed
	}

	if req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidRating
	}

	title := sql.NullString{
		String: "",
		Valid:  false,
	}

	if req.Title != nil && utf8.RuneCountInString(*req.Title) != 0 {
		title.Valid = true
		title.String = *req.Title
	}

	if utf8.RuneCountInString(req.Content) < 10 {
		return nil, ErrReviewContentTooShort
	}

	r := &domain.Review{
		ID:        uuid.NewString(),
		BookID:    bookID,
		UserID:    userID,
		Rating:    req.Rating,
		Title:     title,
		Content:   req.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	uSum := u.ToSummary()
	return r.ToResponse(uSum), s.reviewRepo.Create(ctx, r)
}

func (s *ReviewService) GetByID(ctx context.Context, id string) (*domain.ReviewResponse, error) {
	r, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrReviewNotFound
	}

	u, err := s.userRepo.GetByID(ctx, r.UserID)
	if err != nil {
		return nil, err
	}

	return r.ToResponse(u.ToSummary()), nil
}

func (s *ReviewService) ListByBookID(
	ctx context.Context,
	bookID string,
	page, limit int,
) (*domain.ReviewListResponse, error) {
	if _, err := s.bookRepo.GetByID(ctx, bookID); err != nil {
		return nil, ErrBookNotFound
	}

	list, count, err := s.reviewRepo.ListByBookID(ctx, bookID, page, limit)
	if err != nil {
		return nil, err
	}

	data := make([]domain.ReviewResponse, 0, len(list))
	uIDs := make([]string, 0, len(list))

	for _, v := range list {

		uIDs = append(uIDs, v.UserID)
	}

	for _, v := range list {

		users, err := s.userRepo.GetByIDs(ctx, uIDs)
		if err != nil {
			slog.Error("ReviewService ListByBookID()", "error", err)
			continue
		}

		u := users[v.UserID]
		data = append(data, *v.ToResponse(u.ToSummary()))

	}

	return &domain.ReviewListResponse{
		Data: data,
		Pagination: domain.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      count,
			TotalPages: (count - limit + 1) / limit,
		},
	}, nil
}

func (s *ReviewService) Update(
	ctx context.Context,
	userID, reviewID string,
	req domain.UpdateReviewRequest,
) (*domain.ReviewResponse, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	r, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, ErrReviewNotFound
	}

	if r.UserID != u.ID {
		return nil, ErrNotReviewOwner
	}

	title := r.Title
	if req.Title != nil {
		if utf8.RuneCountInString(*req.Title) == 0 {
			title = sql.NullString{String: "", Valid: false}
		} else {
			title = sql.NullString{String: *req.Title, Valid: true}
		}
	}

	var rating int
	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return nil, ErrInvalidRating
		}

		rating = *req.Rating
	} else {
		rating = r.Rating
	}

	var content string
	if req.Content != nil {
		if utf8.RuneCountInString(*req.Content) < 10 {
			return nil, ErrReviewContentTooShort
		}

		content = *req.Content
	} else {
		content = r.Content
	}

	r.Title = title
	r.Content = content
	r.Rating = rating

	if err = s.reviewRepo.Update(ctx, r); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}

	return r.ToResponse(u.ToSummary()), nil
}

func (s *ReviewService) Delete(
	ctx context.Context,
	userID, reviewID string,
) error {
	r, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return ErrReviewNotFound
	}

	if userID != r.UserID {
		return ErrNotReviewOwner
	}

	return s.reviewRepo.Delete(ctx, reviewID)
}
