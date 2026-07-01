package repository

import (
	"bookshelf/auth-svc/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
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

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return r.getByField(ctx, "id", id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.getByField(ctx, "email", email)
}

func (r *UserRepository) getByField(ctx context.Context, field string, val interface{}) (*domain.User, error) {
	res, err := getEntityByField[domain.User](ctx, r.db, "users", field, val)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return res, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.getByField(ctx, "username", username)
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	q := `UPDATE users SET id = :id, username = :username, email = :email, password_hash = :password_hash, created_at = :created_at, updated_at = :updated_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, q, user)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) bool {
	if _, err := r.GetByEmail(ctx, email); err != nil {
		return false
	}
	return true
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) bool {
	if _, err := r.GetByUsername(ctx, username); err != nil {
		return false
	}
	return true
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []string) (map[string]*domain.User, error) {
	res := make(map[string]*domain.User)
	list := make([]*domain.User, 0)

	q, args, err := sqlx.In(`SELECT (id, username, email, created_at, updated_at) FROM users WHERE id IN (?);`, ids)
	if err != nil {
		return nil, err
	}

	q = r.db.Rebind(q)

	if err = r.db.SelectContext(ctx, &list, q, args...); err != nil {
		return nil, err
	}

	for _, v := range list {
		res[v.ID] = v
	}
	return res, nil
}

func (r *UserRepository) GetByUsernames(ctx context.Context, usernames []string) (map[string]*domain.User, error) {
	res := make(map[string]*domain.User)
	list := make([]*domain.User, 0)

	q, args, err := sqlx.In(`SELECT * FROM users WHERE username IN (?);`, usernames)
	if err != nil {
		return nil, err
	}

	q = r.db.Rebind(q)

	if err = r.db.SelectContext(ctx, &list, q, args...); err != nil {
		return nil, err
	}

	for _, v := range list {
		res[v.Username] = v
	}

	return res, err
}
