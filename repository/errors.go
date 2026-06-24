package repository

import (
	"context"
	"errors"
)

var (
	// ErrNotFound when a product row does not exist.
	ErrNotFound = errors.New("product not found")
	// ErrDuplicateSKU when sku or slug already exists.
	ErrDuplicateSKU = errors.New("duplicate sku or slug")
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
