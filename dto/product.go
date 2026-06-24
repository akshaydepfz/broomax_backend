package dto

import (
	"time"

	"oryoo.com/models"
)

// CreateProductRequest is the body for POST /products.
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=255"`
	SKU         string  `json:"sku" validate:"required,min=2,max=100"`
	ProductCode string  `json:"product_code" validate:"omitempty,max=100"`
	Slug        string  `json:"slug" validate:"omitempty,max=255"`

	CategoryID      string `json:"category_id" validate:"required,uuid"`
	SubCategoryID   string `json:"sub_category_id" validate:"omitempty,uuid"`
	BrandID         string `json:"brand_id" validate:"required,uuid"`

	Description      string `json:"description" validate:"omitempty,max=5000"`
	ShortDescription string `json:"short_description" validate:"omitempty,max=500"`

	DealerPrice float64 `json:"dealer_price" validate:"gte=0"`
	RetailPrice float64 `json:"retail_price" validate:"gte=0"`
	MRP         float64 `json:"mrp" validate:"gte=0"`

	GSTRate float64 `json:"gst_rate" validate:"gte=0,lte=100"`
	HSNCode string  `json:"hsn_code" validate:"omitempty,max=20"`

	StockQty     int `json:"stock_qty" validate:"gte=0"`
	ReorderLevel int `json:"reorder_level" validate:"gte=0"`
	MinOrderQty  int `json:"min_order_qty" validate:"gte=1"`
	MaxOrderQty  int `json:"max_order_qty" validate:"gte=1"`

	Status       string `json:"status" validate:"omitempty,oneof=draft active inactive archived"`
	IsFeatured   bool   `json:"is_featured"`
	IsTrending   bool   `json:"is_trending"`
	IsNewArrival bool   `json:"is_new_arrival"`

	Thumbnail string   `json:"thumbnail" validate:"omitempty"`
	Images    []string `json:"images"`

	CatalogPDF  string `json:"catalog_pdf" validate:"omitempty"`
	WarrantyPDF string `json:"warranty_pdf" validate:"omitempty"`

	CompatibleVehicleBrands []string               `json:"compatible_vehicle_brands"`
	CompatibleModels        []string               `json:"compatible_models"`
	CompatibleYears         []string               `json:"compatible_years"`
	SupplierIDs             []string               `json:"supplier_ids" validate:"omitempty,dive,uuid"`
	WarehouseIDs            []string               `json:"warehouse_ids" validate:"omitempty,dive,uuid"`
	Specifications          []models.ProductSpecification `json:"specifications"`
	Variants                []models.ProductVariant       `json:"variants"`

	MetaTitle       string `json:"meta_title" validate:"omitempty,max=255"`
	MetaDescription string `json:"meta_description" validate:"omitempty,max=500"`
	CreatedBy       string `json:"created_by" validate:"omitempty,max=100"`
}

// UpdateProductRequest is the body for PUT /products/:id.
type UpdateProductRequest struct {
	Name        *string  `json:"name" validate:"omitempty,min=2,max=255"`
	SKU         *string  `json:"sku" validate:"omitempty,min=2,max=100"`
	ProductCode *string  `json:"product_code" validate:"omitempty,max=100"`
	Slug        *string  `json:"slug" validate:"omitempty,max=255"`

	CategoryID    *string `json:"category_id" validate:"omitempty,uuid"`
	SubCategoryID *string `json:"sub_category_id" validate:"omitempty,uuid"`
	BrandID       *string `json:"brand_id" validate:"omitempty,uuid"`

	Description      *string `json:"description" validate:"omitempty,max=5000"`
	ShortDescription *string `json:"short_description" validate:"omitempty,max=500"`

	DealerPrice *float64 `json:"dealer_price" validate:"omitempty,gte=0"`
	RetailPrice *float64 `json:"retail_price" validate:"omitempty,gte=0"`
	MRP         *float64 `json:"mrp" validate:"omitempty,gte=0"`

	GSTRate *float64 `json:"gst_rate" validate:"omitempty,gte=0,lte=100"`
	HSNCode *string  `json:"hsn_code" validate:"omitempty,max=20"`

	StockQty     *int `json:"stock_qty" validate:"omitempty,gte=0"`
	ReorderLevel *int `json:"reorder_level" validate:"omitempty,gte=0"`
	MinOrderQty  *int `json:"min_order_qty" validate:"omitempty,gte=1"`
	MaxOrderQty  *int `json:"max_order_qty" validate:"omitempty,gte=1"`

	Status       *string `json:"status" validate:"omitempty,oneof=draft active inactive archived"`
	IsFeatured   *bool   `json:"is_featured"`
	IsTrending   *bool   `json:"is_trending"`
	IsNewArrival *bool   `json:"is_new_arrival"`

	Thumbnail *string  `json:"thumbnail" validate:"omitempty"`
	Images    []string `json:"images"`

	CatalogPDF  *string `json:"catalog_pdf" validate:"omitempty"`
	WarrantyPDF *string `json:"warranty_pdf" validate:"omitempty"`

	CompatibleVehicleBrands []string                        `json:"compatible_vehicle_brands"`
	CompatibleModels        []string                        `json:"compatible_models"`
	CompatibleYears         []string                        `json:"compatible_years"`
	SupplierIDs             []string                        `json:"supplier_ids" validate:"omitempty,dive,uuid"`
	WarehouseIDs            []string                        `json:"warehouse_ids" validate:"omitempty,dive,uuid"`
	Specifications          []models.ProductSpecification   `json:"specifications"`
	Variants                []models.ProductVariant         `json:"variants"`

	MetaTitle       *string `json:"meta_title" validate:"omitempty,max=255"`
	MetaDescription *string `json:"meta_description" validate:"omitempty,max=500"`
	UpdatedBy       string  `json:"updated_by" validate:"omitempty,max=100"`
}

// BulkUploadRequest is the body for POST /products/bulk-upload.
type BulkUploadRequest struct {
	Products []CreateProductRequest `json:"products" validate:"required,min=1,dive"`
}

// ProductListQuery holds query params for GET /products.
type ProductListQuery struct {
	Search      string `form:"search"`
	CategoryID  string `form:"category_id"`
	BrandID     string `form:"brand_id"`
	SupplierID  string `form:"supplier_id"`
	StockFilter string `form:"stock_filter"` // in_stock | low_stock | out_of_stock
	Status      string `form:"status"`
	SortBy      string `form:"sort_by"`    // name | created_at | dealer_price | stock_qty
	SortOrder   string `form:"sort_order"` // asc | desc
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

// ProductListResponse paginated list payload.
type ProductListResponse struct {
	Items      []models.Product `json:"items"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// BulkUploadResponse result of bulk import.
type BulkUploadResponse struct {
	Created int      `json:"created"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// AddImagesRequest optional JSON body when not using multipart.
type AddImagesRequest struct {
	URLs []string `json:"urls" validate:"required,min=1"`
}

// AddDocumentsRequest optional JSON body when not using multipart.
type AddDocumentsRequest struct {
	Documents []DocumentInput `json:"documents" validate:"omitempty,dive"`
}

// DocumentInput describes a document to attach.
type DocumentInput struct {
	DocType  string `json:"doc_type" validate:"required,oneof=catalog warranty datasheet manual other"`
	URL      string `json:"url" validate:"required"`
	Filename string `json:"filename" validate:"omitempty,max=255"`
}

// ProductResponse wraps a single product for API responses.
type ProductResponse struct {
	models.Product
}

// ToProduct maps a create request into a domain product (IDs/timestamps filled by service).
func (r CreateProductRequest) ToProduct() models.Product {
	status := r.Status
	if status == "" {
		status = models.ProductStatusDraft
	}
	return models.Product{
		Name:                    r.Name,
		SKU:                     r.SKU,
		ProductCode:             r.ProductCode,
		Slug:                    r.Slug,
		CategoryID:              r.CategoryID,
		SubCategoryID:           r.SubCategoryID,
		BrandID:                 r.BrandID,
		Description:             r.Description,
		ShortDescription:        r.ShortDescription,
		DealerPrice:             r.DealerPrice,
		RetailPrice:             r.RetailPrice,
		MRP:                     r.MRP,
		GSTRate:                 r.GSTRate,
		HSNCode:                 r.HSNCode,
		StockQty:                r.StockQty,
		ReorderLevel:            r.ReorderLevel,
		MinOrderQty:             r.MinOrderQty,
		MaxOrderQty:             r.MaxOrderQty,
		Status:                  status,
		IsFeatured:              r.IsFeatured,
		IsTrending:              r.IsTrending,
		IsNewArrival:            r.IsNewArrival,
		Thumbnail:               r.Thumbnail,
		Images:                  r.Images,
		CatalogPDF:              r.CatalogPDF,
		WarrantyPDF:             r.WarrantyPDF,
		CompatibleVehicleBrands: r.CompatibleVehicleBrands,
		CompatibleModels:        r.CompatibleModels,
		CompatibleYears:         r.CompatibleYears,
		SupplierIDs:             r.SupplierIDs,
		WarehouseIDs:            r.WarehouseIDs,
		Specifications:          r.Specifications,
		Variants:                r.Variants,
		MetaTitle:               r.MetaTitle,
		MetaDescription:         r.MetaDescription,
		CreatedBy:               r.CreatedBy,
		UpdatedBy:               r.CreatedBy,
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}
}
