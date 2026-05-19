package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"frontdev333/bookshelf/internal/domain"
	"log/slog"
	"strconv"
	"time"

	"frontdev333/bookshelf/internal/repository"
)

var (
	ErrBookNotFound    = errors.New("book not found")
	ErrNotBookOwner    = errors.New("you are not the owner of this book")
	ErrBookTitleEmpty  = errors.New("book title is empty")
	ErrBookAuthorEmpty = errors.New("book author is empty")
)

type BookService struct {
	bookRepo   *repository.BookRepository
	userRepo   *repository.UserRepository
	reviewRepo *repository.ReviewRepository
}

func (s *BookService) Create(
	ctx context.Context,
	userId string,
	req domain.CreateBookRequest,
) (*domain.BookResponse, error) {
	if req.Title == "" {
		return nil, ErrBookTitleEmpty
	}

	if req.Author == "" {
		return nil, ErrBookAuthorEmpty
	}

	desc := sql.NullString{
		String: *req.Description,
		Valid:  true,
	}

	isbn := sql.NullString{
		String: *req.ISBN,
		Valid:  true,
	}

	pubYear := sql.NullInt32{
		Int32: int32(*req.PublishedYear),
		Valid: true,
	}

	b := &domain.Book{
		Title:         req.Title,
		Author:        req.Author,
		CreatedBy:     userId,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Description:   desc,
		ISBN:          isbn,
		PublishedYear: pubYear,
		AverageRating: sql.NullFloat64{},
	}

	if err := s.bookRepo.Create(ctx, b); err != nil {
		return nil, err
	}

	reviewsCount, err := s.reviewRepo.GetReviewsCount(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	return b.ToResponse(domain.UserSummary{
		ID:       userId,
		Username: u.Username,
	}, &reviewsCount), nil
}

func (s *BookService) GetByID(ctx context.Context, id string) (*domain.BookResponse, error) {
	b, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetByID(ctx, b.CreatedBy)
	if err != nil {
		return nil, err
	}

	reviewsCount, err := s.reviewRepo.GetReviewsCount(ctx, id)
	if err != nil {
		return nil, err
	}

	return b.ToResponse(u.ToSummary(), &reviewsCount), nil
}

func (s *BookService) List(ctx context.Context, f domain.BookFilter) (*domain.BookListResponse, error) {
	f.SeedDefaults()

	var order string
	var sort string
	var search string
	var page int
	var limit int
	var offset int
	var err error

	if f.Order != nil {
		switch *f.Order {
		case "author":
			order = "author"
		case "published_year":
			order = "published_year"
		case "created_by":
			order = "created_by"
		case "updated_at":
			order = "updated_at"
		default:
			order = "created_at"
		}
	}

	if f.Sort == nil || *f.Sort != "DESC" && *f.Sort != "ASC" {
		sort = "DESC"
	} else {
		sort = "ASC"
	}

	if f.Search != nil {
		search = "%" + *f.Search + "%"
	}

	if f.Page != nil {
		page, err = strconv.Atoi(*f.Page)
		if err != nil {
			return nil, err
		}
	} else {
		page = 1
	}

	if f.Limit != nil {
		limit, err = strconv.Atoi(*f.Limit)
		if err != nil {
			return nil, err
		}
	} else {
		limit = 10
	}

	offset = (page - 1) * limit

	//TODO:: проверить

	list, count, err := s.bookRepo.List(ctx, order, sort, search, page, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("s.bookRepo.List(): %w", err)
	}

	booksResponse := make([]domain.BookResponse, 0, len(list))
	authorsNicknames := make([]string, 0)
	booksIDs := make([]string, 0, len(list))

	for _, v := range list {
		authorsNicknames = append(authorsNicknames, v.Author)
		booksIDs = append(booksIDs, v.ID)
	}

	authors, err := s.userRepo.GetByUsernames(ctx, authorsNicknames)
	if err != nil {
		return nil, err
	}

	reviewsCounts, err := s.reviewRepo.GetReviewsCounts(ctx, booksIDs)
	if err != nil {
		return nil, fmt.Errorf("s.reviewRepo.GetReviewsCounts(): %w", err)
	}

	for _, v := range list {
		u, ok := authors[v.Author]
		if !ok || u == nil {
			slog.Error("BookService List()", "error", "author not found", "username", v.Author)
			continue
		}

		reviewsCount, ok := reviewsCounts[v.ID]
		if !ok {
			slog.Warn("reviews count not found", "book_id", v.ID)
			continue
		}
		b := v.ToResponse(u.ToSummary(), &reviewsCount)

		booksResponse = append(booksResponse, *b)
	}

	return &domain.BookListResponse{
		Data: booksResponse,
		Pagination: domain.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      count,
			TotalPages: (count + limit - 1) / limit,
		},
	}, nil
}

func (s *BookService) Update(ctx context.Context, userID, bookID string, req domain.UpdateBookRequest) (*domain.BookResponse, error) {
	b, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, ErrBookNotFound
	}

	if b.CreatedBy != userID {
		return nil, ErrNotBookOwner
	}

	if req.Author != nil {
		b.Author = *req.Author
	}

	if req.ISBN != nil {
		b.ISBN = sql.NullString{
			String: *req.ISBN,
			Valid:  true,
		}
	}

	if req.Description != nil {
		b.Description = sql.NullString{
			String: *req.Description,
			Valid:  true,
		}
	}

	if req.Title != nil {
		b.Title = *req.Title
	}

	if req.PublishedYear != nil {
		b.PublishedYear = sql.NullInt32{
			Int32: int32(*req.PublishedYear),
			Valid: true,
		}
	}

	if err = s.bookRepo.Update(ctx, b); err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	reviewsCount, err := s.reviewRepo.GetReviewsCount(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	return b.ToResponse(u.ToSummary(), &reviewsCount), nil
}

func (s *BookService) Delete(ctx context.Context, userID, bookID string) error {
	b, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return ErrBookNotFound
	}

	if b.CreatedBy != userID {
		return ErrNotBookOwner
	}

	return s.bookRepo.Delete(ctx, b.ID)
}
