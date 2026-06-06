// Package users manages account profiles, registration records, and permissions.
package users

import (
	"context"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/utils"

	"github.com/google/uuid"
)

// UserService defines user account management business operations.
type UserService interface {
	CreateUser(ctx context.Context, email, password string, role Role) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type userService struct {
	repo UserRepository
}

// NewUserService instantiates a new user service mapped to the repository database dependency.
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

// CreateUser coordinates validation checks, secure password hashing, and stores a new user record.
func (s *userService) CreateUser(ctx context.Context, email, password string, role Role) (*User, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.NewInternalError("database error checking email availability", err)
	}
	if existing != nil {
		return nil, errors.NewConflictError("email already in use", nil)
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.NewInternalError("failed to secure user password", err)
	}

	user := &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashed,
		Role:         role,
		Verified:     false,
	}

	err = s.repo.Create(ctx, user)
	if err != nil {
		return nil, errors.NewInternalError("failed to record user in database", err)
	}

	return user, nil
}

// GetUserByID loads a single user by primary key ID, returning a NotFoundError if missing.
func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve user", err)
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user not found", nil)
	}
	return user, nil
}

// GetUserByEmail loads a single user by email address.
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve user by email", err)
	}
	return user, nil
}
