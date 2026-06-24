package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"oryoo.com/dto"
	"oryoo.com/models"
	"oryoo.com/repository"
	"oryoo.com/validation"
)

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// ProductService contains product business logic.
type ProductService struct {
	repo repository.ProductRepository
}

// NewProductService constructs a ProductService.
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// Create validates and persists a new product.
func (s *ProductService) Create(ctx context.Context, req dto.CreateProductRequest) (*models.Product, error) {
	if err := validation.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	if req.MaxOrderQty < req.MinOrderQty {
		return nil, fmt.Errorf("validation: max_order_qty must be >= min_order_qty")
	}

	p := req.ToProduct()
	if err := s.prepareProduct(ctx, &p, ""); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, &p); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, p.ID)
}

// GetByID returns a single product.
func (s *ProductService) GetByID(ctx context.Context, id string) (*models.Product, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("validation: invalid product id")
	}
	return s.repo.GetByID(ctx, id)
}

// List returns paginated, filtered products.
func (s *ProductService) List(ctx context.Context, q dto.ProductListQuery) (*dto.ProductListResponse, error) {
	normalizeListQuery(&q)
	filter := repository.ListFilter{
		Search:      q.Search,
		CategoryID:  q.CategoryID,
		BrandID:     q.BrandID,
		SupplierID:  q.SupplierID,
		StockFilter: q.StockFilter,
		Status:      q.Status,
		SortBy:      q.SortBy,
		SortOrder:   q.SortOrder,
		Page:        q.Page,
		PageSize:    q.PageSize,
	}
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	totalPages := int(math.Ceil(float64(total) / float64(q.PageSize)))
	if total == 0 {
		totalPages = 0
	}
	return &dto.ProductListResponse{
		Items:      items,
		Total:      total,
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Update applies partial updates to a product.
func (s *ProductService) Update(ctx context.Context, id string, req dto.UpdateProductRequest) (*models.Product, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("validation: invalid product id")
	}
	if err := validation.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	applyUpdate(existing, req)

	if existing.MaxOrderQty < existing.MinOrderQty {
		return nil, fmt.Errorf("validation: max_order_qty must be >= min_order_qty")
	}

	if err := s.prepareProduct(ctx, existing, id); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Delete removes a product by id.
func (s *ProductService) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("validation: invalid product id")
	}
	return s.repo.Delete(ctx, id)
}

// BulkUpload creates multiple products; continues on row errors.
func (s *ProductService) BulkUpload(ctx context.Context, req dto.BulkUploadRequest) (*dto.BulkUploadResponse, error) {
	if err := validation.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	resp := &dto.BulkUploadResponse{}
	for i, item := range req.Products {
		if _, err := s.Create(ctx, item); err != nil {
			resp.Failed++
			resp.Errors = append(resp.Errors, fmt.Sprintf("row %d (%s): %v", i+1, item.SKU, err))
			continue
		}
		resp.Created++
	}
	return resp, nil
}

// AddImages attaches image URLs to a product.
func (s *ProductService) AddImages(ctx context.Context, productID string, urls []string) ([]models.ProductImage, error) {
	if _, err := uuid.Parse(productID); err != nil {
		return nil, fmt.Errorf("validation: invalid product id")
	}
	ok, err := s.repo.Exists(ctx, productID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, repository.ErrNotFound
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("validation: at least one image url is required")
	}
	return s.repo.AddImages(ctx, productID, urls)
}

// AddDocuments attaches documents to a product.
func (s *ProductService) AddDocuments(ctx context.Context, productID string, inputs []dto.DocumentInput) ([]models.ProductDocument, error) {
	if _, err := uuid.Parse(productID); err != nil {
		return nil, fmt.Errorf("validation: invalid product id")
	}
	req := dto.AddDocumentsRequest{Documents: inputs}
	if err := validation.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	ok, err := s.repo.Exists(ctx, productID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, repository.ErrNotFound
	}
	var docs []models.ProductDocument
	for _, in := range inputs {
		docs = append(docs, models.ProductDocument{
			DocType:  in.DocType,
			URL:      in.URL,
			Filename: in.Filename,
		})
	}
	return s.repo.AddDocuments(ctx, productID, docs)
}

func (s *ProductService) prepareProduct(ctx context.Context, p *models.Product, excludeID string) error {
	p.SKU = strings.TrimSpace(strings.ToUpper(p.SKU))
	p.Name = strings.TrimSpace(p.Name)
	p.HSNCode = strings.TrimSpace(p.HSNCode)

	if p.Slug == "" {
		p.Slug = slugify(p.Name)
	} else {
		p.Slug = slugify(p.Slug)
	}

	exists, err := s.repo.SKUExists(ctx, p.SKU, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return repository.ErrDuplicateSKU
	}
	exists, err = s.repo.SlugExists(ctx, p.Slug, excludeID)
	if err != nil {
		return err
	}
	if exists {
		p.Slug = p.Slug + "-" + strings.ToLower(p.SKU)
	}

	catName, err := s.repo.ResolveCategoryName(ctx, p.CategoryID)
	if err != nil {
		return err
	}
	p.CategoryName = catName

	subName, err := s.repo.ResolveSubCategoryName(ctx, p.SubCategoryID)
	if err != nil {
		return err
	}
	p.SubCategoryName = subName

	brandName, err := s.repo.ResolveBrandName(ctx, p.BrandID)
	if err != nil {
		return err
	}
	p.BrandName = brandName

	for i := range p.Variants {
		if p.Variants[i].ID == "" {
			p.Variants[i].ID = uuid.NewString()
		}
		p.Variants[i].SKU = strings.TrimSpace(strings.ToUpper(p.Variants[i].SKU))
	}

	if p.MinOrderQty == 0 {
		p.MinOrderQty = 1
	}
	if p.MaxOrderQty == 0 {
		p.MaxOrderQty = 9999
	}
	if p.Status == "" {
		p.Status = models.ProductStatusDraft
	}
	return nil
}

func applyUpdate(p *models.Product, req dto.UpdateProductRequest) {
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.SKU != nil {
		p.SKU = *req.SKU
	}
	if req.ProductCode != nil {
		p.ProductCode = *req.ProductCode
	}
	if req.Slug != nil {
		p.Slug = *req.Slug
	}
	if req.CategoryID != nil {
		p.CategoryID = *req.CategoryID
	}
	if req.SubCategoryID != nil {
		p.SubCategoryID = *req.SubCategoryID
	}
	if req.BrandID != nil {
		p.BrandID = *req.BrandID
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.ShortDescription != nil {
		p.ShortDescription = *req.ShortDescription
	}
	if req.DealerPrice != nil {
		p.DealerPrice = *req.DealerPrice
	}
	if req.RetailPrice != nil {
		p.RetailPrice = *req.RetailPrice
	}
	if req.MRP != nil {
		p.MRP = *req.MRP
	}
	if req.GSTRate != nil {
		p.GSTRate = *req.GSTRate
	}
	if req.HSNCode != nil {
		p.HSNCode = *req.HSNCode
	}
	if req.StockQty != nil {
		p.StockQty = *req.StockQty
	}
	if req.ReorderLevel != nil {
		p.ReorderLevel = *req.ReorderLevel
	}
	if req.MinOrderQty != nil {
		p.MinOrderQty = *req.MinOrderQty
	}
	if req.MaxOrderQty != nil {
		p.MaxOrderQty = *req.MaxOrderQty
	}
	if req.Status != nil {
		p.Status = *req.Status
	}
	if req.IsFeatured != nil {
		p.IsFeatured = *req.IsFeatured
	}
	if req.IsTrending != nil {
		p.IsTrending = *req.IsTrending
	}
	if req.IsNewArrival != nil {
		p.IsNewArrival = *req.IsNewArrival
	}
	if req.Thumbnail != nil {
		p.Thumbnail = *req.Thumbnail
	}
	if req.Images != nil {
		p.Images = req.Images
	}
	if req.CatalogPDF != nil {
		p.CatalogPDF = *req.CatalogPDF
	}
	if req.WarrantyPDF != nil {
		p.WarrantyPDF = *req.WarrantyPDF
	}
	if req.CompatibleVehicleBrands != nil {
		p.CompatibleVehicleBrands = req.CompatibleVehicleBrands
	}
	if req.CompatibleModels != nil {
		p.CompatibleModels = req.CompatibleModels
	}
	if req.CompatibleYears != nil {
		p.CompatibleYears = req.CompatibleYears
	}
	if req.SupplierIDs != nil {
		p.SupplierIDs = req.SupplierIDs
	}
	if req.WarehouseIDs != nil {
		p.WarehouseIDs = req.WarehouseIDs
	}
	if req.Specifications != nil {
		p.Specifications = req.Specifications
	}
	if req.Variants != nil {
		p.Variants = req.Variants
	}
	if req.MetaTitle != nil {
		p.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		p.MetaDescription = *req.MetaDescription
	}
	if req.UpdatedBy != "" {
		p.UpdatedBy = req.UpdatedBy
	}
}

func normalizeListQuery(q *dto.ProductListQuery) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
	if q.SortOrder == "" {
		q.SortOrder = "desc"
	}
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return uuid.NewString()
	}
	return s
}

// MapServiceError converts repository errors to HTTP-friendly messages.
func MapServiceError(err error) (status int, msg string) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return 404, "product not found"
	case errors.Is(err, repository.ErrDuplicateSKU):
		return 409, "sku or slug already exists"
	case errors.Is(err, repository.ErrInvalidReference):
		return 400, "invalid category, subcategory, or brand reference"
	default:
		if err != nil && strings.HasPrefix(err.Error(), "validation:") {
			return 400, strings.TrimPrefix(err.Error(), "validation: ")
		}
		return 500, "internal server error"
	}
}
