package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors for common domain error cases.
var (
	ErrNotFound       = errors.New("entity not found")
	ErrOptimisticLock = errors.New("optimistic lock conflict: entity was modified by another transaction")
	ErrInvalidState   = errors.New("invalid state transition")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrAlreadyExists  = errors.New("entity already exists")
	ErrValidation     = errors.New("validation error")
)

// DomainError is a structured error with code and message for API responses.
type DomainError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Detail     string `json:"detail,omitempty"`
	Err        error  `json:"-"`
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a 404 domain error.
func NewNotFoundError(entity, id string) *DomainError {
	return &DomainError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s not found", entity),
		Detail:  fmt.Sprintf("%s with id '%s' was not found", entity, id),
		Err:     ErrNotFound,
	}
}

// NewValidationError creates a 400 domain error.
func NewValidationError(message string) *DomainError {
	return &DomainError{
		Code:    http.StatusBadRequest,
		Message: message,
		Err:     ErrValidation,
	}
}

// NewConflictError creates a 409 domain error.
func NewConflictError(message string) *DomainError {
	return &DomainError{
		Code:    http.StatusConflict,
		Message: message,
		Err:     ErrOptimisticLock,
	}
}

// NewInvalidStateError creates a 422 domain error.
func NewInvalidStateError(from, to string) *DomainError {
	return &DomainError{
		Code:    http.StatusUnprocessableEntity,
		Message: fmt.Sprintf("cannot transition from '%s' to '%s'", from, to),
		Err:     ErrInvalidState,
	}
}

// NewUnauthorizedError creates a 401 domain error.
func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Code:    http.StatusUnauthorized,
		Message: message,
		Err:     ErrUnauthorized,
	}
}

// NewForbiddenError creates a 403 domain error.
func NewForbiddenError(message string) *DomainError {
	return &DomainError{
		Code:    http.StatusForbidden,
		Message: message,
		Err:     ErrForbidden,
	}
}

// NewAlreadyExistsError creates a 409 domain error for duplicates.
func NewAlreadyExistsError(entity, field, value string) *DomainError {
	return &DomainError{
		Code:    http.StatusConflict,
		Message: fmt.Sprintf("%s with %s '%s' already exists", entity, field, value),
		Err:     ErrAlreadyExists,
	}
}
