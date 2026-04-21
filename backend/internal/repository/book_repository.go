package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"frontdev333/bookshelf/internal/domain"
	"strconv"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type bookRepository struct {
	db *sqlx.DB
}

func (r *bookRepository) Create(ctx context.Context, book *domain.Book) error {
	q := `INSERT INTO books (id, title, author, description, isbn, published_year, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	bookID := uuid.NewString()
	if _, err := r.db.ExecContext(ctx, q, bookID, book.Title, book.Author, book.Description, book.ISBN, book.PublishedYear, book.CreatedBy, book.CreatedAt, book.UpdatedAt); err != nil {
		return err
	}
	book.ID = bookID
	return nil
}

func (r *bookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	var b domain.Book
	q := `
		SELECT books.*, avg(r.rating) AS avg_rate FROM books
			LEFT JOIN reviews as r ON r.book_id = books.id
		WHERE books.id = $1
		GROUP BY books.id 
		LIMIT 1
`
	if err := r.db.GetContext(ctx, &b, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &b, nil
}

func (r *bookRepository) List(ctx context.Context, f domain.BookFilter) ([]domain.Book, int, error) {
	qList := fmt.Sprintf(`SELECT * FROM books ORDER BY %s %s LIMIT $3 OFFSET $4`, f.Order, f.Sort)
	qCount := `SELECT COUNT(*) FROM books`
	var res []domain.Book

	var page int
	var limit int
	var offset int
	var err error

	if f.Page != nil {
		page, err = strconv.Atoi(*f.Page)
		if err != nil {
			return nil, 0, err
		}
	} else {
		page = 1
	}

	if f.Limit != nil {

		limit, err = strconv.Atoi(*f.Limit)
		if err != nil {
			return nil, 0, err
		}
	} else {
		limit = 10
	}

	offset = (page - 1) * limit

	err = r.db.SelectContext(ctx, &res, qList, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	var count int
	if err = r.db.SelectContext(ctx, &count, qCount); err != nil {
		return nil, 0, err
	}

	return res, count, nil
}

func (r *bookRepository) Update(ctx context.Context, book *domain.Book) error {
	q := `UPDATE books SET id = :id, title = :title, author = :author, description = :description, isbn = :isbn, published_year = :published_year, created_by = :created_by, created_at = :created_at, updated_at = :updated_at WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, q, book); err != nil {
		return err
	}
	return nil
}

func (r *bookRepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM books WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, q, id); err != nil {
		return err
	}
	return nil
}
