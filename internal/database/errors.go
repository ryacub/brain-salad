// Package database provides custom error types for type-safe error handling.
package database

import "errors"

// Sentinel errors for common database operations
var (
	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists indicates a resource with the same identifier already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput indicates the provided input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrConstraintViolation indicates a database constraint was violated
	ErrConstraintViolation = errors.New("constraint violation")
)

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if an error is an "already exists" error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInput checks if an error is an "invalid input" error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsConstraintViolation checks if an error is a "constraint violation" error
func IsConstraintViolation(err error) bool {
	return errors.Is(err, ErrConstraintViolation)
}
