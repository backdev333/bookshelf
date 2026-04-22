package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("not found")
var ErrUserAlreadyExists = errors.New("user already exists")

type Repository struct {
	User   *UserRepository
	Book   *bookRepository
	Review *reviewRepository
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		User:   &UserRepository{db},
		Book:   &bookRepository{db},
		Review: &reviewRepository{db},
	}
}

func getEntityByField[T any](ctx context.Context, db *sqlx.DB, table, field string, val interface{}) (*T, error) {
	var entity T

	q := `SELECT * FROM ` + table + ` WHERE ` + field + ` = $1 LIMIT 1`
	err := db.GetContext(ctx, &entity, q, val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}
