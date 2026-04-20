package repository

import (
	"context"
	"frontdev333/bookshelf/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type reviewRepository struct {
	db *sqlx.DB
}

func (r *reviewRepository) Create(ctx context.Context, review *domain.Review) error {
	reviewID := uuid.NewString()

	q := `INSERT INTO reviews (id, book_id, user_id, rating, title, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	if _, err := r.db.ExecContext(ctx, q, reviewID, review.BookID, review.UserID, review.Rating, review.Title, review.Content, review.CreatedAt, review.UpdatedAt); err != nil {
		return err
	}
	review.ID = reviewID
	return nil
}

func (r *reviewRepository) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	return r.getByField(ctx, "id", id)
}

func (r *reviewRepository) getByField(ctx context.Context, field string, val interface{}) (*domain.Review, error) {
	return getEntityByField[domain.Review](ctx, r.db, "reviews", field, val)
}

func (r *reviewRepository) ListByBookID(ctx context.Context, bookID string, page, limit int) ([]domain.Review, int, error) {
	var res []domain.Review

	q := `SELECT * FROM books LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, q, limit, page*limit)
	defer rows.Close()

	if err != nil {
		return nil, 0, err
	}

	for rows.Next() {
		var review domain.Review
		if err = rows.Scan(&review); err != nil {
			return nil, 0, err
		}

		res = append(res, review)
	}

	if rows.Err() != nil {
		return nil, 0, err
	}

	return res, len(res), nil
}

func (r *reviewRepository) Update(ctx context.Context, review *domain.Review) error {
	q := `UPDATE reviews SET id = :id, book_id = :book_id, user_id = :user_id, rating = :rating, title = :title, content = :content, created_at = :created_at, updated_at = :updated_at WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, q, review); err != nil {
		return err
	}
	return nil
}

func (r *reviewRepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM reviews WHERE id = :id`
	if _, err := r.db.ExecContext(ctx, q, id); err != nil {
		return err
	}
	return nil
}

func (r *reviewRepository) UserHasReviewedBook(ctx context.Context, userID, bookID string) (bool, error) {
	var res bool
	q := `SELECT EXISTS(SELECT 1 FROM reviews WHERE user_id = $1 AND book_id = $2 LIMIT 1)`
	if err := r.db.QueryRowContext(ctx, q, userID, bookID).Scan(&res); err != nil {
		return false, err
	}
	return res, nil
}
