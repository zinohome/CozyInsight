package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE username = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &user, query, username); err != nil {
		return nil, fmt.Errorf("find user by username failed: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		return nil, fmt.Errorf("find user by id failed: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (username, password_hash, email, nick_name, avatar, phone, status, is_admin)
			  VALUES (:username, :password_hash, :email, :nick_name, :avatar, :phone, :status, :is_admin)`
	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	id, _ := result.LastInsertId()
	user.ID = uint64(id)
	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uint64) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("update last login failed: %w", err)
	}
	return nil
}
