package service

import (
	service2 "bookshelf/auth-service/internal/service"
	"frontdev333/bookshelf/internal/repository"
)

type Service struct {
	User   *service2.UserService
	Book   *BookService
	Review *ReviewService
}

func New(repos *repository.Repository, jwtSecret string) *Service {
	return &Service{
		User:   &service2.UserService{repo: repos.User, jwtSecret: jwtSecret},
		Book:   &BookService{bookRepo: repos.Book, userRepo: repos.User, reviewRepo: repos.Review},
		Review: &ReviewService{repos.Review, repos.Book, repos.User},
	}
}
