package repository

import (
	"context"

	"oryoo.com/models"
)

// ProductRepository defines persistence operations for products.
type ProductRepository interface {
	Create(ctx context.Context, p *models.Product) error
	GetByID(ctx context.Context, id string) (*models.Product, error)
	List(ctx context.Context, f ListFilter) ([]models.Product, int, error)
	Update(ctx context.Context, p *models.Product) error
	Delete(ctx context.Context, id string) error
	SKUExists(ctx context.Context, sku string, excludeID string) (bool, error)
	SlugExists(ctx context.Context, slug string, excludeID string) (bool, error)
	ResolveCategoryName(ctx context.Context, categoryID string) (string, error)
	ResolveSubCategoryName(ctx context.Context, subCategoryID string) (string, error)
	ResolveBrandName(ctx context.Context, brandID string) (string, error)
	AddImages(ctx context.Context, productID string, urls []string) ([]models.ProductImage, error)
	AddDocuments(ctx context.Context, productID string, docs []models.ProductDocument) ([]models.ProductDocument, error)
	Exists(ctx context.Context, id string) (bool, error)
}
