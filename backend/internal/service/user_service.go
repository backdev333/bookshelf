package service

import (
	"context"
	"errors"
	"fmt"
	"frontdev333/bookshelf/internal/domain"
	"frontdev333/bookshelf/internal/repository"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserExists = errors.New("user already exists")
var ErrInvalidPassword = errors.New("invalid password")
var ErrInvalidUsername = errors.New("invalid username")
var ErrInvalidEmail = errors.New("invalid email")
var ErrUsernameExists = errors.New("username already exists")

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func (s *UserService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, ErrInvalidEmail
	}

	u, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:        u.ToPublic(),
		AccessToken: token,
	}, nil
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

	token, err := s.generateToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:        u.ToPublic(),
		AccessToken: token,
	}, nil
}

func (s *UserService) ValidateToken(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(s.jwtSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return "", fmt.Errorf("validate token: %w", err)
	}

	if claims.Subject == "" {
		return "", errors.New("token subject is empty")
	}

	return claims.Subject, nil
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
		return ErrUsernameExists
	}
	return nil
}

func (s *UserService) generateToken(userID string) (string, error) {

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
