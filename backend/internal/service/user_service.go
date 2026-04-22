package service

import (
	"context"
	"errors"
	"frontdev333/bookshelf/internal/domain"
	"frontdev333/bookshelf/internal/repository"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserExists = errors.New("user already exists")
var ErrInvalidPassword = errors.New("invalid password")
var ErrInvalidUsername = errors.New("invalid username")
var ErrInvalidEmail = errors.New("invalid email")

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func (s *UserService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {

	if err := s.validateRegisterReq(ctx, req); err != nil {
		return nil, err
	}

	passBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passBytes),
	}

	if err = s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:        u.ToPublic(),
		AccessToken: "",
	}, nil
}

func (s *UserService) validateRegisterReq(ctx context.Context, req domain.RegisterRequest) error {
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return ErrInvalidEmail
	}

	if len([]byte(req.Username)) < 3 {
		return ErrInvalidUsername
	}

	if len([]byte(req.Password)) < 8 {
		return ErrInvalidPassword
	}

	if s.repo.EmailExists(ctx, req.Email) {
		return ErrUserExists
	}

	if u, err := s.repo.GetByUsername(ctx, req.Username); err == nil && u != nil {
		return ErrUserExists
	}
	return nil
}
