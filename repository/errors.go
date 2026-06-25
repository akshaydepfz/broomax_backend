package repository

import (
	"errors"
)

var (
	// ErrNotFound when a row does not exist.
	ErrNotFound = errors.New("not found")
	// ErrDuplicateSKU when sku or slug already exists.
	ErrDuplicateSKU = errors.New("duplicate sku or slug")
	// ErrDuplicateSlug when a category slug already exists.
	ErrDuplicateSlug = errors.New("duplicate slug")
	// ErrCategoryInUse when a category is referenced by products or subcategories.
	ErrCategoryInUse = errors.New("category in use")
	// ErrInvalidReference when category/brand/subcategory FK is invalid.
	ErrInvalidReference = errors.New("invalid category, subcategory, or brand reference")
)

// ListFilter drives search, pagination, sorting, and filters for product listing.
type ListFilter struct {
	Search      string
	CategoryID  string
	BrandID     string
	SupplierID  string
	StockFilter string
	Status      string
	SortBy      string
	SortOrder   string
	Page        int
	PageSize    int
}
