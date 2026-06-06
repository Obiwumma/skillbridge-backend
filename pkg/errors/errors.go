// Package errors defines unified custom application errors mapped to HTTP response states.
package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a structured, categorised domain error for API consumer responses.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

// Error formats the AppError into a readable string representation.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAppError is the base constructor for compiling custom AppErrors.
func NewAppError(httpStatus int, code, message string, err error) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

// NewValidationError builds an HTTP 400 validation error for corrupt payloads.
func NewValidationError(message string, err error) *AppError {
	return NewAppError(http.StatusBadRequest, "VALIDATION_ERROR", message, err)
}

// NewUnauthorizedError builds an HTTP 401 error for missing or invalid sessions.
func NewUnauthorizedError(message string, err error) *AppError {
	return NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", message, err)
}

// NewForbiddenError builds an HTTP 403 error when access limits are breached.
func NewForbiddenError(message string, err error) *AppError {
	return NewAppError(http.StatusForbidden, "FORBIDDEN", message, err)
}

// NewNotFoundError builds an HTTP 404 error for missing database records.
func NewNotFoundError(message string, err error) *AppError {
	return NewAppError(http.StatusNotFound, "NOT_FOUND", message, err)
}

// NewConflictError builds an HTTP 409 error when records violate unique constraints.
func NewConflictError(message string, err error) *AppError {
	return NewAppError(http.StatusConflict, "CONFLICT", message, err)
}

// NewInternalError builds an HTTP 500 unhandled system failure error.
func NewInternalError(message string, err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, err)
}
