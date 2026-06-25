package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"oryoo.com/models"
)

// PostgresCategoryRepository persists categories in PostgreSQL.
type PostgresCategoryRepository struct {
	db *sql.DB
}

// NewPostgresCategoryRepository constructs a PostgresCategoryRepository.
func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

var _ CategoryRepository = (*PostgresCategoryRepository)(nil)

const categorySelectCols = `
	id, name, image_url, description, sort_order, is_active, created_at, updated_at
`

func scanCategory(row interface {
	Scan(dest ...any) error
}) (*models.Category, error) {
	var c models.Category
	if err := row.Scan(
		&c.ID,
		&c.Title,
		&c.ImageURL,
		&c.Description,
		&c.SortOrder,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &c, nil
}

// Create inserts a new category row.
func (r *PostgresCategoryRepository) Create(ctx context.Context, c *models.Category, slug string) error {
	q := `
		INSERT INTO categories (name, slug, image_url, description, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	row := r.db.QueryRowContext(ctx, q,
		c.Title, slug, c.ImageURL, c.Description, c.SortOrder, c.IsActive,
	)
	if err := row.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if isPgUniqueViolation(err) {
			return ErrDuplicateSlug
		}
		return fmt.Errorf("create category: %w", err)
	}
	return nil
}

// GetByID returns one category by primary key.
func (r *PostgresCategoryRepository) GetByID(ctx context.Context, id string) (*models.Category, error) {
	q := `SELECT ` + categorySelectCols + ` FROM categories WHERE id = $1`
	c, err := scanCategory(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	return c, nil
}

// List returns paginated categories matching the filter.
func (r *PostgresCategoryRepository) List(ctx context.Context, f CategoryListFilter) ([]models.Category, int, error) {
	where, args := buildCategoryListWhere(f)
	countQ := `SELECT COUNT(*) FROM categories` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count categories: %w", err)
	}

	sortCol := categorySortColumn(f.SortBy)
	sortDir := "ASC"
	if strings.EqualFold(f.SortOrder, "desc") {
		sortDir = "DESC"
	}

	offset := (f.Page - 1) * f.PageSize
	listArgs := append(append([]any{}, args...), f.PageSize, offset)
	q := `SELECT ` + categorySelectCols + ` FROM categories` + where +
		fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortCol, sortDir, len(args)+1, len(args)+2)

	rows, err := r.db.QueryContext(ctx, q, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var items []models.Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan category: %w", err)
		}
		items = append(items, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if items == nil {
		items = []models.Category{}
	}
	return items, total, nil
}

// Update modifies an existing category.
func (r *PostgresCategoryRepository) Update(ctx context.Context, c *models.Category, slug string) error {
	q := `
		UPDATE categories
		SET name = $1, slug = $2, image_url = $3, description = $4,
		    sort_order = $5, is_active = $6, updated_at = now()
		WHERE id = $7
		RETURNING created_at, updated_at
	`
	row := r.db.QueryRowContext(ctx, q,
		c.Title, slug, c.ImageURL, c.Description, c.SortOrder, c.IsActive, c.ID,
	)
	if err := row.Scan(&c.CreatedAt, &c.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if isPgUniqueViolation(err) {
			return ErrDuplicateSlug
		}
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

// Delete removes a category by id.
func (r *PostgresCategoryRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		if isPgForeignKeyViolation(err) {
			return ErrCategoryInUse
		}
		return fmt.Errorf("delete category: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// SlugExists reports whether a slug is already taken.
func (r *PostgresCategoryRepository) SlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM categories WHERE slug = $1`
	args := []any{slug}
	if excludeID != "" {
		q += ` AND id <> $2`
		args = append(args, excludeID)
	}
	q += `)`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func buildCategoryListWhere(f CategoryListFilter) (string, []any) {
	var clauses []string
	var args []any
	n := 1

	if strings.TrimSpace(f.Search) != "" {
		clauses = append(clauses, fmt.Sprintf("lower(name) LIKE $%d", n))
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(f.Search))+"%")
		n++
	}
	if f.IsActive != nil {
		clauses = append(clauses, fmt.Sprintf("is_active = $%d", n))
		args = append(args, *f.IsActive)
		n++
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func categorySortColumn(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "title", "name":
		return "name"
	case "created_at":
		return "created_at"
	case "updated_at":
		return "updated_at"
	default:
		return "sort_order"
	}
}
