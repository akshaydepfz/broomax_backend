package repository

import (
	"context"

	"oryoo.com/models"
)

// CategoryListFilter drives category listing.
type CategoryListFilter struct {
	Search    string
	IsActive  *bool
	SortBy    string
	SortOrder string
	Page      int
	PageSize  int
}

// CategoryRepository defines persistence operations for categories.
type CategoryRepository interface {
	Create(ctx context.Context, c *models.Category, slug string) error
	GetByID(ctx context.Context, id string) (*models.Category, error)
	List(ctx context.Context, f CategoryListFilter) ([]models.Category, int, error)
	Update(ctx context.Context, c *models.Category, slug string) error
	Delete(ctx context.Context, id string) error
	SlugExists(ctx context.Context, slug string, excludeID string) (bool, error)
}
