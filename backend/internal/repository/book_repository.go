package repository

import (
	"context"
	"frontdev333/bookshelf/internal/domain"

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
			JOIN reviews as r ON r.book_id = books.id
		WHERE id = $1
		GROUP BY r.book_id 
		LIMIT 1
`
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&b); err != nil {
		return nil, err
	}

	return &b, nil
}

func (r *bookRepository) List(ctx context.Context, f domain.BookFilter) ([]domain.Book, int, error) {
	qList := `SELECT * FROM books ORDER BY $1 $2 LIMIT $3 OFFSET $4`
	qCount := `SELECT COUNT(*) FROM books`
	var res []domain.Book
	err := r.db.SelectContext(ctx, &res, qList, f.Order, f.Sort, f.Limit, f.Page)
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
	q := `DELETE FROM books WHERE id = :id`
	if _, err := r.db.ExecContext(ctx, q, id); err != nil {
		return err
	}
	return nil
}
