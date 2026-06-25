package models

import "time"

// ProductStatus values for catalog lifecycle.
const (
	ProductStatusDraft     = "draft"
	ProductStatusActive    = "active"
	ProductStatusInactive  = "inactive"
	ProductStatusArchived  = "archived"
)

// Product is the domain model aligned with the Flutter ProductModel.
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SKU         string `json:"sku"`
	ProductCode string `json:"product_code"`
	Slug        string `json:"slug"`

	CategoryID      string `json:"category_id"`
	CategoryName    string `json:"category_name"`
	SubCategoryID   string `json:"sub_category_id"`
	SubCategoryName string `json:"sub_category_name"`
	BrandID         string `json:"brand_id"`
	BrandName       string `json:"brand_name"`

	Description      string `json:"description"`
	ShortDescription string `json:"short_description"`

	DealerPrice float64 `json:"dealer_price"`
	RetailPrice float64 `json:"retail_price"`
	MRP         float64 `json:"mrp"`

	GSTRate float64 `json:"gst_rate"`
	HSNCode string  `json:"hsn_code"`

	StockQty     int `json:"stock_qty"`
	ReorderLevel int `json:"reorder_level"`
	MinOrderQty  int `json:"min_order_qty"`
	MaxOrderQty  int `json:"max_order_qty"`

	Status       string `json:"status"`
	IsFeatured   bool   `json:"is_featured"`
	IsTrending   bool   `json:"is_trending"`
	IsNewArrival bool   `json:"is_new_arrival"`

	Thumbnail string   `json:"thumbnail"`
	Images    []string `json:"images"`

	CatalogPDF  string `json:"catalog_pdf"`
	WarrantyPDF string `json:"warranty_pdf"`

	CompatibleVehicleBrands []string `json:"compatible_vehicle_brands"`
	CompatibleModels        []string `json:"compatible_models"`
	CompatibleYears         []string `json:"compatible_years"`

	SupplierIDs   []string `json:"supplier_ids"`
	WarehouseIDs  []string `json:"warehouse_ids"`
	Specifications []ProductSpecification `json:"specifications"`
	Variants       []ProductVariant       `json:"variants"`

	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
}

// ProductImage is a persisted product image record.
type ProductImage struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	URL       string    `json:"url"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// ProductDocument is a persisted product document record.
type ProductDocument struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	DocType   string    `json:"doc_type"`
	URL       string    `json:"url"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
}

// SubCategory reference entity.
type SubCategory struct {
	ID         string `json:"id"`
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
}

// Brand reference entity.
type Brand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
