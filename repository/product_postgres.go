package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"oryoo.com/models"
)

// PostgresProductRepository implements ProductRepository for PostgreSQL.
type PostgresProductRepository struct {
	db *sql.DB
}

// NewPostgresProductRepository returns a PostgreSQL-backed product repository.
func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *models.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	specJSON, err := json.Marshal(p.Specifications)
	if err != nil {
		return err
	}
	variantsJSON, err := json.Marshal(p.Variants)
	if err != nil {
		return err
	}
	cvbJSON, err := json.Marshal(stringSliceOrEmpty(p.CompatibleVehicleBrands))
	if err != nil {
		return err
	}
	cmJSON, err := json.Marshal(stringSliceOrEmpty(p.CompatibleModels))
	if err != nil {
		return err
	}
	cyJSON, err := json.Marshal(stringSliceOrEmpty(p.CompatibleYears))
	if err != nil {
		return err
	}

	var subCat any
	if p.SubCategoryID != "" {
		subCat = p.SubCategoryID
	}
	var category any
	if p.CategoryID != "" {
		category = p.CategoryID
	}
	var brand any
	if p.BrandID != "" {
		brand = p.BrandID
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO products (
			name, sku, product_code, slug,
			category_id, sub_category_id, brand_id,
			description, short_description,
			dealer_price, retail_price, mrp,
			gst_rate, hsn_code,
			stock_qty, reorder_level, min_order_qty, max_order_qty,
			status, is_featured, is_trending, is_new_arrival,
			thumbnail, catalog_pdf, warranty_pdf,
			compatible_vehicle_brands, specifications, variants,
			compatible_models, compatible_years,
			meta_title, meta_description,
			created_by, updated_by
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,
			$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34
		) RETURNING id, created_at, updated_at`,
		p.Name, p.SKU, p.ProductCode, p.Slug,
		category, subCat, brand,
		p.Description, p.ShortDescription,
		p.DealerPrice, p.RetailPrice, p.MRP,
		p.GSTRate, p.HSNCode,
		p.StockQty, p.ReorderLevel, p.MinOrderQty, p.MaxOrderQty,
		p.Status, p.IsFeatured, p.IsTrending, p.IsNewArrival,
		p.Thumbnail, p.CatalogPDF, p.WarrantyPDF,
		cvbJSON, specJSON, variantsJSON, cmJSON, cyJSON,
		p.MetaTitle, p.MetaDescription,
		p.CreatedBy, p.UpdatedBy,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isPgUniqueViolation(err) {
			return ErrDuplicateSKU
		}
		if isPgForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return err
	}

	if err := r.syncImagesTx(ctx, tx, p.ID, p.Images); err != nil {
		return err
	}
	if err := r.syncSuppliersTx(ctx, tx, p.ID, p.SupplierIDs); err != nil {
		return err
	}
	if err := r.syncWarehousesTx(ctx, tx, p.ID, p.WarehouseIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			p.id, p.name, p.sku, p.product_code, p.slug,
			p.category_id, COALESCE(c.name, ''),
			p.sub_category_id, COALESCE(sc.name, ''),
			p.brand_id, COALESCE(b.name, ''),
			p.description, p.short_description,
			p.dealer_price, p.retail_price, p.mrp,
			p.gst_rate, p.hsn_code,
			p.stock_qty, p.reorder_level, p.min_order_qty, p.max_order_qty,
			p.status, p.is_featured, p.is_trending, p.is_new_arrival,
			p.thumbnail, p.catalog_pdf, p.warranty_pdf,
			p.compatible_vehicle_brands, p.specifications, p.variants,
			p.compatible_models, p.compatible_years,
			p.meta_title, p.meta_description,
			p.created_at, p.updated_at, p.created_by, p.updated_by
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN subcategories sc ON sc.id = p.sub_category_id
		LEFT JOIN brands b ON b.id = p.brand_id
		WHERE p.id = $1`, id)

	p, err := scanProduct(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	images, err := r.loadImages(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Images = images

	suppliers, err := r.loadSuppliers(ctx, id)
	if err != nil {
		return nil, err
	}
	p.SupplierIDs = suppliers

	warehouses, err := r.loadWarehouses(ctx, id)
	if err != nil {
		return nil, err
	}
	p.WarehouseIDs = warehouses

	return p, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, f ListFilter) ([]models.Product, int, error) {
	where, args := buildListWhere(f)
	sortCol := mapSortColumn(f.SortBy)
	sortDir := "DESC"
	if strings.EqualFold(f.SortOrder, "asc") {
		sortDir = "ASC"
	}

	countSQL := `
		SELECT COUNT(DISTINCT p.id)
		FROM products p
		LEFT JOIN product_suppliers ps ON ps.product_id = p.id
		WHERE 1=1` + where

	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PageSize
	listArgs := append(append([]any{}, args...), f.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT
			p.id, p.name, p.sku, p.product_code, p.slug,
			p.category_id, COALESCE(c.name, ''),
			p.sub_category_id, COALESCE(sc.name, ''),
			p.brand_id, COALESCE(b.name, ''),
			p.description, p.short_description,
			p.dealer_price, p.retail_price, p.mrp,
			p.gst_rate, p.hsn_code,
			p.stock_qty, p.reorder_level, p.min_order_qty, p.max_order_qty,
			p.status, p.is_featured, p.is_trending, p.is_new_arrival,
			p.thumbnail, p.catalog_pdf, p.warranty_pdf,
			p.compatible_vehicle_brands, p.specifications, p.variants,
			p.compatible_models, p.compatible_years,
			p.meta_title, p.meta_description,
			p.created_at, p.updated_at, p.created_by, p.updated_by
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN subcategories sc ON sc.id = p.sub_category_id
		LEFT JOIN brands b ON b.id = p.brand_id
		LEFT JOIN product_suppliers ps ON ps.product_id = p.id
		WHERE 1=1`+where+`
		ORDER BY p.`+sortCol+` `+sortDir+`
		LIMIT $`+fmt.Sprint(len(listArgs)-1)+` OFFSET $`+fmt.Sprint(len(listArgs)), listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		images, err := r.loadImages(ctx, p.ID)
		if err != nil {
			return nil, 0, err
		}
		p.Images = images
		suppliers, err := r.loadSuppliers(ctx, p.ID)
		if err != nil {
			return nil, 0, err
		}
		p.SupplierIDs = suppliers
		warehouses, err := r.loadWarehouses(ctx, p.ID)
		if err != nil {
			return nil, 0, err
		}
		p.WarehouseIDs = warehouses
		items = append(items, *p)
	}
	return items, total, rows.Err()
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *models.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	specJSON, _ := json.Marshal(p.Specifications)
	variantsJSON, _ := json.Marshal(p.Variants)
	cvbJSON, _ := json.Marshal(stringSliceOrEmpty(p.CompatibleVehicleBrands))
	cmJSON, _ := json.Marshal(stringSliceOrEmpty(p.CompatibleModels))
	cyJSON, _ := json.Marshal(stringSliceOrEmpty(p.CompatibleYears))

	var subCat any
	if p.SubCategoryID != "" {
		subCat = p.SubCategoryID
	}
	var category any
	if p.CategoryID != "" {
		category = p.CategoryID
	}
	var brand any
	if p.BrandID != "" {
		brand = p.BrandID
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE products SET
			name=$1, sku=$2, product_code=$3, slug=$4,
			category_id=$5, sub_category_id=$6, brand_id=$7,
			description=$8, short_description=$9,
			dealer_price=$10, retail_price=$11, mrp=$12,
			gst_rate=$13, hsn_code=$14,
			stock_qty=$15, reorder_level=$16, min_order_qty=$17, max_order_qty=$18,
			status=$19, is_featured=$20, is_trending=$21, is_new_arrival=$22,
			thumbnail=$23, catalog_pdf=$24, warranty_pdf=$25,
			compatible_vehicle_brands=$26, specifications=$27, variants=$28,
			compatible_models=$29, compatible_years=$30,
			meta_title=$31, meta_description=$32,
			updated_by=$33, updated_at=now()
		WHERE id=$34`,
		p.Name, p.SKU, p.ProductCode, p.Slug,
		category, subCat, brand,
		p.Description, p.ShortDescription,
		p.DealerPrice, p.RetailPrice, p.MRP,
		p.GSTRate, p.HSNCode,
		p.StockQty, p.ReorderLevel, p.MinOrderQty, p.MaxOrderQty,
		p.Status, p.IsFeatured, p.IsTrending, p.IsNewArrival,
		p.Thumbnail, p.CatalogPDF, p.WarrantyPDF,
		cvbJSON, specJSON, variantsJSON, cmJSON, cyJSON,
		p.MetaTitle, p.MetaDescription,
		p.UpdatedBy, p.ID,
	)
	if err != nil {
		if isPgUniqueViolation(err) {
			return ErrDuplicateSKU
		}
		if isPgForeignKeyViolation(err) {
			return ErrInvalidReference
		}
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	if err := r.syncImagesTx(ctx, tx, p.ID, p.Images); err != nil {
		return err
	}
	if err := r.syncSuppliersTx(ctx, tx, p.ID, p.SupplierIDs); err != nil {
		return err
	}
	if err := r.syncWarehousesTx(ctx, tx, p.ID, p.WarehouseIDs); err != nil {
		return err
	}

	if err := tx.QueryRowContext(ctx, `SELECT updated_at FROM products WHERE id=$1`, p.ID).Scan(&p.UpdatedAt); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresProductRepository) SKUExists(ctx context.Context, sku string, excludeID string) (bool, error) {
	var exists bool
	q := `SELECT EXISTS(SELECT 1 FROM products WHERE lower(sku)=lower($1)`
	args := []any{sku}
	if excludeID != "" {
		q += ` AND id <> $2`
		args = append(args, excludeID)
	}
	q += `)`
	return exists, r.db.QueryRowContext(ctx, q, args...).Scan(&exists)
}

func (r *PostgresProductRepository) SlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	var exists bool
	q := `SELECT EXISTS(SELECT 1 FROM products WHERE slug=$1`
	args := []any{slug}
	if excludeID != "" {
		q += ` AND id <> $2`
		args = append(args, excludeID)
	}
	q += `)`
	return exists, r.db.QueryRowContext(ctx, q, args...).Scan(&exists)
}

func (r *PostgresProductRepository) ResolveCategoryName(ctx context.Context, categoryID string) (string, error) {
	if categoryID == "" {
		return "", nil
	}
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM categories WHERE id=$1`, categoryID).Scan(&name)
	if err == sql.ErrNoRows {
		return "", ErrInvalidReference
	}
	return name, err
}

func (r *PostgresProductRepository) ResolveSubCategoryName(ctx context.Context, subCategoryID string) (string, error) {
	if subCategoryID == "" {
		return "", nil
	}
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM subcategories WHERE id=$1`, subCategoryID).Scan(&name)
	if err == sql.ErrNoRows {
		return "", ErrInvalidReference
	}
	return name, err
}

func (r *PostgresProductRepository) ResolveBrandName(ctx context.Context, brandID string) (string, error) {
	if brandID == "" {
		return "", nil
	}
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM brands WHERE id=$1`, brandID).Scan(&name)
	if err == sql.ErrNoRows {
		return "", ErrInvalidReference
	}
	return name, err
}

func (r *PostgresProductRepository) AddImages(ctx context.Context, productID string, urls []string) ([]models.ProductImage, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var maxOrder int
	_ = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort_order), -1) FROM product_images WHERE product_id=$1`, productID).Scan(&maxOrder)

	var out []models.ProductImage
	for i, url := range urls {
		var img models.ProductImage
		order := maxOrder + i + 1
		err := tx.QueryRowContext(ctx, `
			INSERT INTO product_images (product_id, url, sort_order)
			VALUES ($1, $2, $3)
			RETURNING id, product_id, url, sort_order, created_at`,
			productID, url, order,
		).Scan(&img.ID, &img.ProductID, &img.URL, &img.SortOrder, &img.CreatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, img)
	}

	if len(out) > 0 && maxOrder == -1 {
		_, _ = tx.ExecContext(ctx, `UPDATE products SET thumbnail=$1, updated_at=now() WHERE id=$2 AND thumbnail=''`, out[0].URL, productID)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PostgresProductRepository) AddDocuments(ctx context.Context, productID string, docs []models.ProductDocument) ([]models.ProductDocument, error) {
	var out []models.ProductDocument
	for _, d := range docs {
		var saved models.ProductDocument
		err := r.db.QueryRowContext(ctx, `
			INSERT INTO product_documents (product_id, doc_type, url, filename)
			VALUES ($1, $2, $3, $4)
			RETURNING id, product_id, doc_type, url, filename, created_at`,
			productID, d.DocType, d.URL, d.Filename,
		).Scan(&saved.ID, &saved.ProductID, &saved.DocType, &saved.URL, &saved.Filename, &saved.CreatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, saved)

		switch d.DocType {
		case "catalog":
			_, _ = r.db.ExecContext(ctx, `UPDATE products SET catalog_pdf=$1, updated_at=now() WHERE id=$2`, d.URL, productID)
		case "warranty":
			_, _ = r.db.ExecContext(ctx, `UPDATE products SET warranty_pdf=$1, updated_at=now() WHERE id=$2`, d.URL, productID)
		}
	}
	return out, nil
}

func (r *PostgresProductRepository) Exists(ctx context.Context, id string) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id=$1)`, id).Scan(&ok)
	return ok, err
}

// --- helpers ---

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(s rowScanner) (*models.Product, error) {
	var p models.Product
	var categoryID, subCatID, brandID sql.NullString
	var cvb, spec, variants, cm, cy []byte

	err := s.Scan(
		&p.ID, &p.Name, &p.SKU, &p.ProductCode, &p.Slug,
		&categoryID, &p.CategoryName,
		&subCatID, &p.SubCategoryName,
		&brandID, &p.BrandName,
		&p.Description, &p.ShortDescription,
		&p.DealerPrice, &p.RetailPrice, &p.MRP,
		&p.GSTRate, &p.HSNCode,
		&p.StockQty, &p.ReorderLevel, &p.MinOrderQty, &p.MaxOrderQty,
		&p.Status, &p.IsFeatured, &p.IsTrending, &p.IsNewArrival,
		&p.Thumbnail, &p.CatalogPDF, &p.WarrantyPDF,
		&cvb, &spec, &variants, &cm, &cy,
		&p.MetaTitle, &p.MetaDescription,
		&p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	if categoryID.Valid {
		p.CategoryID = categoryID.String
	}
	if subCatID.Valid {
		p.SubCategoryID = subCatID.String
	}
	if brandID.Valid {
		p.BrandID = brandID.String
	}
	_ = json.Unmarshal(cvb, &p.CompatibleVehicleBrands)
	_ = json.Unmarshal(spec, &p.Specifications)
	_ = json.Unmarshal(variants, &p.Variants)
	_ = json.Unmarshal(cm, &p.CompatibleModels)
	_ = json.Unmarshal(cy, &p.CompatibleYears)
	return &p, nil
}

func (r *PostgresProductRepository) loadImages(ctx context.Context, productID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT url FROM product_images WHERE product_id=$1 ORDER BY sort_order, created_at`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, rows.Err()
}

func (r *PostgresProductRepository) loadSuppliers(ctx context.Context, productID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT supplier_id FROM product_suppliers WHERE product_id=$1`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStringRows(rows)
}

func (r *PostgresProductRepository) loadWarehouses(ctx context.Context, productID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT warehouse_id FROM product_warehouses WHERE product_id=$1`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStringRows(rows)
}

func scanStringRows(rows *sql.Rows) ([]string, error) {
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *PostgresProductRepository) syncImagesTx(ctx context.Context, tx *sql.Tx, productID string, urls []string) error {
	if urls == nil {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM product_images WHERE product_id=$1`, productID); err != nil {
		return err
	}
	for i, url := range urls {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO product_images (product_id, url, sort_order) VALUES ($1, $2, $3)`,
			productID, url, i); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresProductRepository) syncSuppliersTx(ctx context.Context, tx *sql.Tx, productID string, ids []string) error {
	if ids == nil {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM product_suppliers WHERE product_id=$1`, productID); err != nil {
		return err
	}
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO product_suppliers (product_id, supplier_id) VALUES ($1, $2)`,
			productID, id); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresProductRepository) syncWarehousesTx(ctx context.Context, tx *sql.Tx, productID string, ids []string) error {
	if ids == nil {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM product_warehouses WHERE product_id=$1`, productID); err != nil {
		return err
	}
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO product_warehouses (product_id, warehouse_id) VALUES ($1, $2)`,
			productID, id); err != nil {
			return err
		}
	}
	return nil
}

func buildListWhere(f ListFilter) (string, []any) {
	var parts []string
	var args []any
	n := 1

	if s := strings.TrimSpace(f.Search); s != "" {
		parts = append(parts, fmt.Sprintf(`(
			p.name ILIKE $%d OR p.sku ILIKE $%d OR p.product_code ILIKE $%d
			OR p.description ILIKE $%d OR p.hsn_code ILIKE $%d
		)`, n, n, n, n, n))
		args = append(args, "%"+s+"%")
		n++
	}
	if f.CategoryID != "" {
		parts = append(parts, fmt.Sprintf(`p.category_id = $%d`, n))
		args = append(args, f.CategoryID)
		n++
	}
	if f.BrandID != "" {
		parts = append(parts, fmt.Sprintf(`p.brand_id = $%d`, n))
		args = append(args, f.BrandID)
		n++
	}
	if f.SupplierID != "" {
		parts = append(parts, fmt.Sprintf(`ps.supplier_id = $%d`, n))
		args = append(args, f.SupplierID)
		n++
	}
	if f.Status != "" {
		parts = append(parts, fmt.Sprintf(`p.status = $%d`, n))
		args = append(args, f.Status)
		n++
	}
	switch strings.ToLower(f.StockFilter) {
	case "in_stock":
		parts = append(parts, `p.stock_qty > 0`)
	case "low_stock":
		parts = append(parts, `p.stock_qty > 0 AND p.stock_qty <= p.reorder_level`)
	case "out_of_stock":
		parts = append(parts, `p.stock_qty = 0`)
	}

	where := ""
	if len(parts) > 0 {
		where = " AND " + strings.Join(parts, " AND ")
	}
	return where, args
}

func mapSortColumn(sortBy string) string {
	switch strings.ToLower(sortBy) {
	case "name":
		return "name"
	case "dealer_price":
		return "dealer_price"
	case "stock_qty":
		return "stock_qty"
	default:
		return "created_at"
	}
}

func stringSliceOrEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func isPgUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}

func isPgForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23503"
	}
	return strings.Contains(strings.ToLower(err.Error()), "foreign key")
}

// Ensure compile-time interface satisfaction.
var _ ProductRepository = (*PostgresProductRepository)(nil)
