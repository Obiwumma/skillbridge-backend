// Package users manages account profiles, registration records, and permissions.
package users

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserRepository specifies data store behaviors for mapping user records.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}

type sqlUserRepository struct {
	db *sql.DB
}

// NewUserRepository returns a database access client mapped to the PostgreSQL pool.
func NewUserRepository(db *sql.DB) UserRepository {
	return &sqlUserRepository{db: db}
}

// Create records a new user row in the database.
func (r *sqlUserRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Verified,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

// GetByID extracts a single user row matching the UUID.
func (r *sqlUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, password_hash, role, verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail extracts a single user row matching the email string.
func (r *sqlUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, role, verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update mutates email, password hashes, and statuses for an active user.
func (r *sqlUserRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, role = $3, verified = $4, updated_at = $5
		WHERE id = $6
	`
	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Verified,
		user.UpdatedAt,
		user.ID,
	)
	return err
}
