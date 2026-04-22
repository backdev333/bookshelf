package service

import "frontdev333/bookshelf/internal/repository"

type Service struct {
	User   *userService
	Book   *bookService
	Review *reviewService
}

func New(repos *repository.Repository, jwtSecret string) *Service {
	return &Service{
		User:   &UserService{repo: repos.User, jwtSecret: jwtSecret},
		Book:   &BookService{repos.Book},
		Review: &ReviewService{repos.Book},
	}
}
