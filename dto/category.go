package dto

import "oryoo.com/models"

// CreateCategoryRequest is the JSON body for POST /categories.
type CreateCategoryRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	ImageURL    string `json:"image_url" validate:"omitempty,url"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	SortOrder   *int   `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive    *bool  `json:"is_active"`
}

// UpdateCategoryRequest is the JSON body for PUT /categories/:id.
type UpdateCategoryRequest struct {
	Title       *string `json:"title" validate:"omitempty,min=1,max=255"`
	ImageURL    *string `json:"image_url" validate:"omitempty,url"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	SortOrder   *int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive    *bool   `json:"is_active"`
}

// CategoryListQuery drives GET /categories query parameters.
type CategoryListQuery struct {
	Search    string `form:"search"`
	IsActive  *bool  `form:"is_active"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// CategoryListResponse is the paginated list payload.
type CategoryListResponse struct {
	Items      []models.Category `json:"items"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// AddCategoryImageRequest allows setting image_url via JSON.
type AddCategoryImageRequest struct {
	ImageURL string `json:"image_url" validate:"required,url"`
}
