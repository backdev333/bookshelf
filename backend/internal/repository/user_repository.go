package repository

import (
	"context"
	"frontdev333/bookshelf/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

type userRepository struct {
	db *sqlx.DB
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	q := `INSERT INTO users (id, username, email, password_hash) VALUES ($1,$2,$3,$4)`
	userID := uuid.NewString()
	_, err := r.db.ExecContext(ctx, q, userID, user.Username, user.Email, user.PasswordHash)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == pqerror.UniqueViolation {
				return ErrUserAlreadyExists
			}
		}
		return err
	}
	user.ID = userID
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return r.getByField(ctx, "id", id)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.getByField(ctx, "email", email)
}

func (r *userRepository) getByField(ctx context.Context, field string, val interface{}) (*domain.User, error) {
	return getEntityByField[domain.User](ctx, r.db, "users", field, val)
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.getByField(ctx, "username", username)
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	q := `UPDATE users SET id = :id, username = :username, email = :email, password_hash = :password_hash, created_at = :created_at, updated_at = :updated_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, q, user)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) EmailExists(ctx context.Context, email string) bool {
	if _, err := r.GetByEmail(ctx, email); err != nil {
		return false
	}
	return true
}

func (r *userRepository) UsernameExists(ctx context.Context, username string) bool {
	if _, err := r.GetByUsername(ctx, username); err != nil {
		return false
	}
	return true
}
