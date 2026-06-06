// Package users manages account profiles, registration records, and permissions.
package users

import (
	"time"

	"github.com/google/uuid"
)

// Role defines custom string restrictions for role-based system authorization.
type Role string

// Allowed system roles.
const (
	RoleStudent   Role = "student"
	RoleRecruiter Role = "recruiter"
	RoleAdmin     Role = "admin"
)

// User represents a system account record stored in the database.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	Verified     bool      `json:"verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
