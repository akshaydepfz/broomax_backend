package models

// ProductVariant represents a purchasable SKU variant (stored as JSONB on products).
type ProductVariant struct {
	ID          string  `json:"id"`
	VariantName string  `json:"variant_name"`
	SKU         string  `json:"sku"`
	DealerPrice float64 `json:"dealer_price"`
	RetailPrice float64 `json:"retail_price"`
	StockQty    int     `json:"stock_qty"`
	IsDefault   bool    `json:"is_default"`
}
