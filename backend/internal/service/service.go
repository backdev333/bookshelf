package service

import "frontdev333/bookshelf/internal/repository"

type Service struct {
	User   *UserService
	Book   *BookService
	Review *ReviewService
}

func New(repos *repository.Repository, jwtSecret string) *Service {
	return &Service{
		User:   &UserService{repo: repos.User, jwtSecret: jwtSecret},
		Book:   &BookService{bookRepo: repos.Book, userRepo: repos.User, reviewRepo: repos.Review},
		Review: &ReviewService{repos.Review, repos.Book, repos.User},
	}
}
