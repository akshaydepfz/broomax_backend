package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"oryoo.com/dto"
	"oryoo.com/models"
	"oryoo.com/repository"
	"oryoo.com/validation"
)

// CategoryService contains category business logic.
type CategoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService constructs a CategoryService.
func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// Create validates and persists a new category.
func (s *CategoryService) Create(ctx context.Context, req dto.CreateCategoryRequest) (*models.Category, error) {
	if err := validation.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	c := models.Category{
		Title:       strings.TrimSpace(req.Title),
		ImageURL:    strings.TrimSpace(req.ImageURL),
		Description: strings.TrimSpace(req.Description),
		SortOrder:   0,
		IsActive:    true,
	}
	if req.SortOrder != nil {
		c.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}

	slug, err := s.ensureUniqueSlug(ctx, slugify(c.Title), "")
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, &c, slug); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, c.ID)
}

// GetByID returns a single category.
func (s *CategoryService) GetByID(ctx context.Context, id string) (*models.Category, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("validation: invalid category id")
	}
	return s.repo.GetByID(ctx, id)
}

// List returns paginated categories.
func (s *CategoryService) List(ctx context.Context, q dto.CategoryListQuery) (*dto.CategoryListResponse, error) {
	normalizeCategoryListQuery(&q)
	filter := repository.CategoryListFilter{
		Search:    q.Search,
		IsActive:  q.IsActive,
		SortBy:    q.SortBy,
		SortOrder: q.SortOrder,
		Page:      q.Page,
		PageSize:  q.PageSize,
	}
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	totalPages := int(math.Ceil(float64(total) / float64(q.PageSize)))
	if total == 0 {
		totalPages = 0
	}
	return &dto.CategoryListResponse{
		Items:      items,
		Total:      total,
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Update applies partial changes to a category.
func (s *CategoryService) Update(ctx context.Context, id string, req dto.UpdateCategoryRequest) (*models.Category, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("validation: invalid category id")
	}
	if err := validation.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c := *existing
	slug := slugify(c.Title)

	if req.Title != nil {
		c.Title = strings.TrimSpace(*req.Title)
		slug = slugify(c.Title)
	}
	if req.ImageURL != nil {
		c.ImageURL = strings.TrimSpace(*req.ImageURL)
	}
	if req.Description != nil {
		c.Description = strings.TrimSpace(*req.Description)
	}
	if req.SortOrder != nil {
		c.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}

	slug, err = s.ensureUniqueSlug(ctx, slug, id)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, &c, slug); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// SetImageURL updates only the category image URL.
func (s *CategoryService) SetImageURL(ctx context.Context, id string, imageURL string) (*models.Category, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("validation: invalid category id")
	}
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return nil, fmt.Errorf("validation: image_url is required")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	existing.ImageURL = imageURL
	slug := slugify(existing.Title)
	slug, err = s.ensureUniqueSlug(ctx, slug, id)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, existing, slug); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Delete removes a category.
func (s *CategoryService) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("validation: invalid category id")
	}
	return s.repo.Delete(ctx, id)
}

func (s *CategoryService) ensureUniqueSlug(ctx context.Context, base string, excludeID string) (string, error) {
	slug := base
	for i := 0; i < 100; i++ {
		exists, err := s.repo.SlugExists(ctx, slug, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i+2)
	}
	return "", repository.ErrDuplicateSlug
}

func normalizeCategoryListQuery(q *dto.CategoryListQuery) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	if q.SortBy == "" {
		q.SortBy = "sort_order"
	}
	if q.SortOrder == "" {
		q.SortOrder = "asc"
	}
}

// MapCategoryServiceError converts category errors to HTTP-friendly messages.
func MapCategoryServiceError(err error) (status int, msg string) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return 404, "category not found"
	case errors.Is(err, repository.ErrDuplicateSlug):
		return 409, "category title already exists"
	case errors.Is(err, repository.ErrCategoryInUse):
		return 409, "category is in use and cannot be deleted"
	default:
		if err != nil && strings.HasPrefix(err.Error(), "validation:") {
			return 400, strings.TrimPrefix(err.Error(), "validation: ")
		}
		return 500, "internal server error"
	}
}
