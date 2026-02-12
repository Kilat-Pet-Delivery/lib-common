package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository is a generic interface for aggregate persistence.
type Repository[T any] interface {
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
	Save(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// Specification defines a query filter pattern.
type Specification interface {
	ToSQL() (string, []interface{})
}

// PaginatedResult holds a page of results with total count.
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginatedResult creates a paginated result.
func NewPaginatedResult[T any](items []T, total int64, page, limit int) PaginatedResult[T] {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return PaginatedResult[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
