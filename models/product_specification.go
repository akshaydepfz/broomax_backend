package models

// ProductSpecification is a key/value product attribute (stored as JSONB on products).
type ProductSpecification struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
