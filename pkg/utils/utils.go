// Package utils provides stateless cryptographic utility helpers.
package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword secure-hashes a password string utilizing the Bcrypt hashing algorithm.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a raw password against its stored Bcrypt hash value.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
