package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"frontdev333/bookshelf/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrBookNotFound = errors.New("book not found")

type BookRepository struct {
	db *sqlx.DB
}

func (r *BookRepository) Create(ctx context.Context, book *domain.Book) error {
	q := `INSERT INTO books (id, title, author, description, isbn, published_year, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	bookID := uuid.NewString()
	if _, err := r.db.ExecContext(ctx, q, bookID, book.Title, book.Author, book.Description, book.ISBN, book.PublishedYear, book.CreatedBy, book.CreatedAt, book.UpdatedAt); err != nil {
		return err
	}
	book.ID = bookID
	return nil
}

func (r *BookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	var b domain.Book
	q := `
		SELECT books.*, avg(r.rating) AS average_rating FROM books
			LEFT JOIN reviews as r ON r.book_id = books.id
		WHERE books.id = $1
		GROUP BY books.id 
		LIMIT 1
`
	if err := r.db.GetContext(ctx, &b, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}

	return &b, nil
}

func (r *BookRepository) List(
	ctx context.Context,
	order,
	sort,
	search string,
	page,
	limit,
	offset int,

) ([]domain.Book, int, error) {
	var res []domain.Book
	var count int

	// TODO:: не используется page

	rawQ := `
	SELECT id, title, author, description, isbn, published_year, created_by, created_at, updated_at
	FROM books
	WHERE title LIKE $1
	OR description LIKE $2
    ORDER BY %s %s LIMIT $3 OFFSET $4
    `

	qList := fmt.Sprintf(rawQ, order, sort)

	qCount := `SELECT COUNT(*) FROM books WHERE title LIKE $1 OR description LIKE $2`

	if err := r.db.SelectContext(ctx, &res, qList, search, search, limit, offset); err != nil {
		return nil, 0, err
	}

	if err := r.db.GetContext(ctx, &count, qCount, search, search); err != nil {
		return nil, 0, err
	}

	return res, count, nil
}

func (r *BookRepository) Update(ctx context.Context, book *domain.Book) error {
	q := `UPDATE books SET id = :id, title = :title, author = :author, description = :description, isbn = :isbn, published_year = :published_year, created_by = :created_by, created_at = :created_at, updated_at = :updated_at WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, q, book); err != nil {
		return err
	}
	return nil
}

func (r *BookRepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM books WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, q, id); err != nil {
		return err
	}
	return nil
}
