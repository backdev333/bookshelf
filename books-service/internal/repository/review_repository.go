package repository

import (
	"bookshelf/books-service/internal/domain"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ReviewRepository struct {
	db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	reviewID := uuid.NewString()

	q := `INSERT INTO reviews (id, book_id, user_id, rating, title, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	if _, err := r.db.ExecContext(ctx, q, reviewID, review.BookID, review.UserID, review.Rating, review.Title, review.Content, review.CreatedAt, review.UpdatedAt); err != nil {
		return err
	}
	review.ID = reviewID
	return nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	return r.getByField(ctx, "id", id)
}

func (r *ReviewRepository) getByField(ctx context.Context, field string, val interface{}) (*domain.Review, error) {
	return getEntityByField[domain.Review](ctx, r.db, "reviews", field, val)
}

func (r *ReviewRepository) ListByBookID(ctx context.Context, bookID string) ([]domain.Review, error) {
	var res []domain.Review

	qList := `SELECT * FROM reviews WHERE book_id = $1`

	if err := r.db.SelectContext(ctx, &res, qList, bookID); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *ReviewRepository) Update(ctx context.Context, review *domain.Review) error {
	q := `UPDATE reviews SET book_id = :book_id, user_id = :user_id, rating = :rating, title = :title, content = :content, updated_at = :updated_at WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, q, review); err != nil {
		return err
	}

	return nil
}

func (r *ReviewRepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM reviews WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, q, id); err != nil {
		return err
	}
	return nil
}

func (r *ReviewRepository) UserHasReviewedBook(ctx context.Context, userID, bookID string) (bool, error) {
	var res bool
	q := `SELECT EXISTS(SELECT 1 FROM reviews WHERE user_id = $1 AND book_id = $2 LIMIT 1)`
	if err := r.db.QueryRowContext(ctx, q, userID, bookID).Scan(&res); err != nil {
		return false, err
	}
	return res, nil
}

func (r *ReviewRepository) GetReviewsCount(ctx context.Context, bookID string) (int, error) {
	q := `SELECT COUNT(*) FROM reviews WHERE book_id = $1`
	var res int
	if err := r.db.GetContext(ctx, &res, q, bookID); err != nil {
		return 0, err
	}

	return res, nil
}

func (r *ReviewRepository) GetReviewsCounts(ctx context.Context, bookIDs []string) (map[string]int, error) {
	q := `SELECT book_id, COUNT(*) FROM reviews WHERE book_id IN (?) GROUP BY book_id`
	q, args, err := sqlx.In(q, bookIDs)
	if err != nil {
		return nil, err
	}

	q = r.db.Rebind(q)

	res := make([]struct {
		BookID string `db:"book_id"`
		Count  int    `db:"count"`
	}, 0, len(bookIDs))

	if err = r.db.SelectContext(ctx, &res, q, args...); err != nil {
		return nil, err
	}

	toReturn := make(map[string]int, len(res))

	for _, v := range res {
		toReturn[v.BookID] = v.Count
	}

	return toReturn, nil
}
