package service

import (
	"bookshelf/auth-service/internal/domain"
	"bookshelf/auth-service/internal/repository"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrUsernameExists     = errors.New("username already exists")
)

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func NewUserService(repo *repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
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

	token, err := s.generateAccessToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:        u.ToPublic(),
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Now().Add(time.Hour * 24).Unix()),
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

	token, err := s.generateAccessToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:        u.ToPublic(),
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Now().Add(time.Hour * 24).Unix()),
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

	if err := s.validateEmail(ctx, req.Email); err != nil {
		return err
	}

	if err := s.validateUsername(ctx, req.Username); err != nil {
		return err
	}

	if len([]byte(req.Password)) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

func (s *UserService) generateAccessToken(userID string) (string, error) {

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

func (s *UserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *UserService) Update(ctx context.Context, userID string, req domain.UpdateUserRequest) (*domain.User, error) {
	u, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	email := u.Email
	nick := u.Username
	pass := u.PasswordHash

	if req.Username != nil && *req.Username != u.Username {

		if err = s.validateUsername(ctx, *req.Username); err != nil {
			return nil, err
		}

		nick = *req.Username
	}

	if req.Email != nil && *req.Email != u.Email {

		if err = s.validateEmail(ctx, *req.Email); err != nil {
			return nil, err
		}

		email = *req.Email
	}

	if req.Password != nil && bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(*req.Password)) != nil {

		if len([]byte(*req.Password)) < 8 {
			return nil, ErrInvalidPassword
		}

		passBytes, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("password hash", "error", err)
			return nil, ErrInvalidPassword
		}

		pass = string(passBytes)
	}

	res := &domain.User{
		ID:           u.ID,
		Username:     nick,
		Email:        email,
		PasswordHash: pass,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    time.Now(),
	}

	return u, s.repo.Update(ctx, res)
}

func (s *UserService) validateEmail(ctx context.Context, email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}

	if s.repo.EmailExists(ctx, email) {
		return ErrUserExists
	}
	return nil
}

func (s *UserService) validateUsername(ctx context.Context, userName string) error {
	if len([]byte(userName)) < 3 {
		return ErrInvalidUsername
	}

	if s.repo.UsernameExists(ctx, userName) {
		return ErrUsernameExists
	}
	return nil
}
