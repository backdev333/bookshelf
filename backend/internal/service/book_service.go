package service

import (
	"context"
	"database/sql"
	"errors"
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
	bookRepo *repository.BookRepository
	userRepo *repository.UserRepository
}

func (s *BookService) Create(ctx context.Context, userId string, req domain.CreateBookRequest) (*domain.BookResponse, error) {
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

	uSum := &domain.UserSummary{
		ID:       userId,
		Username: req.Author,
	}

	return b.ToResponse(uSum), nil
}

func (s *BookService) GetByID(ctx context.Context, id string) (*domain.BookResponse, error) {
	b, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetByUsername(ctx, b.Author)
	if err != nil {
		return nil, err
	}

	return b.ToResponse(&domain.UserSummary{
		ID:       u.ID,
		Username: u.Username,
	}), nil
}

func (s *BookService) List(ctx context.Context, filter domain.BookFilter) (*domain.BookListResponse, error) {
	filter.SeedDefaults()

	list, count, err := s.bookRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	booksResponse := make([]*domain.BookResponse, len(list))

	for i, v := range list {
		u, err := s.userRepo.GetByUsername(ctx, v.Author)
		if err != nil {
			slog.Error("BookService List()", "error", err)
			continue
		}

		b := v.ToResponse(&domain.UserSummary{
			ID:       u.ID,
			Username: u.Username,
		})

		booksResponse[i] = b
	}

	cPage, err := strconv.Atoi(*filter.Page)
	if err != nil {
		return nil, err
	}

	cLimit, err := strconv.Atoi(*filter.Limit)
	if err != nil {
		return nil, err
	}

	return &domain.BookListResponse{
		Books:       booksResponse,
		Count:       count,
		CurrentPage: cPage,
		Limit:       cLimit,
	}, err
}

func (s *BookService) Update(ctx context.Context, userID, bookID string, req domain.UpdateBookRequest) (*domain.BookResponse, error) {
	b, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
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

	return b.ToResponse(&domain.UserSummary{
		ID:       u.ID,
		Username: u.Username,
	}), nil
}

func (s *BookService) Delete(ctx context.Context, userID, bookID string) error {
	b, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return err
	}

	if b.CreatedBy != userID {
		return ErrNotBookOwner
	}

	return s.bookRepo.Delete(ctx, b.ID)
}
