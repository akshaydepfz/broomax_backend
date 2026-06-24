-- Broomax Automotive B2B Marketplace — product catalog schema (PostgreSQL)
-- Requires: PostgreSQL 14+

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Reference tables
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS subcategories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (category_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_subcategories_category_id ON subcategories(category_id);

CREATE TABLE IF NOT EXISTS brands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Core product table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    sku TEXT NOT NULL UNIQUE,
    product_code TEXT NOT NULL DEFAULT '',
    slug TEXT NOT NULL UNIQUE,

    category_id UUID NOT NULL REFERENCES categories(id),
    sub_category_id UUID REFERENCES subcategories(id),
    brand_id UUID NOT NULL REFERENCES brands(id),

    description TEXT NOT NULL DEFAULT '',
    short_description TEXT NOT NULL DEFAULT '',

    dealer_price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    retail_price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    mrp NUMERIC(12, 2) NOT NULL DEFAULT 0,

    gst_rate NUMERIC(5, 2) NOT NULL DEFAULT 0,
    hsn_code TEXT NOT NULL DEFAULT '',

    stock_qty INT NOT NULL DEFAULT 0,
    reorder_level INT NOT NULL DEFAULT 0,
    min_order_qty INT NOT NULL DEFAULT 1,
    max_order_qty INT NOT NULL DEFAULT 9999,

    status TEXT NOT NULL DEFAULT 'draft',
    is_featured BOOLEAN NOT NULL DEFAULT false,
    is_trending BOOLEAN NOT NULL DEFAULT false,
    is_new_arrival BOOLEAN NOT NULL DEFAULT false,

    thumbnail TEXT NOT NULL DEFAULT '',
    catalog_pdf TEXT NOT NULL DEFAULT '',
    warranty_pdf TEXT NOT NULL DEFAULT '',

    compatible_vehicle_brands JSONB NOT NULL DEFAULT '[]'::jsonb,
    specifications JSONB NOT NULL DEFAULT '[]'::jsonb,
    variants JSONB NOT NULL DEFAULT '[]'::jsonb,
    compatible_models JSONB NOT NULL DEFAULT '[]'::jsonb,
    compatible_years JSONB NOT NULL DEFAULT '[]'::jsonb,

    meta_title TEXT NOT NULL DEFAULT '',
    meta_description TEXT NOT NULL DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by TEXT NOT NULL DEFAULT '',
    updated_by TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_brand_id ON products(brand_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
CREATE INDEX IF NOT EXISTS idx_products_name_lower ON products (lower(name));

-- Gallery images
CREATE TABLE IF NOT EXISTS product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images(product_id);

-- Documents (catalog, warranty, datasheets)
CREATE TABLE IF NOT EXISTS product_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    doc_type TEXT NOT NULL,
    url TEXT NOT NULL,
    filename TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_product_documents_product_id ON product_documents(product_id);

-- Supplier mapping (supplier master assumed external)
CREATE TABLE IF NOT EXISTS product_suppliers (
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL,
    PRIMARY KEY (product_id, supplier_id)
);

CREATE INDEX IF NOT EXISTS idx_product_suppliers_supplier_id ON product_suppliers(supplier_id);

-- Warehouse mapping (warehouse master assumed external)
CREATE TABLE IF NOT EXISTS product_warehouses (
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    warehouse_id UUID NOT NULL,
    PRIMARY KEY (product_id, warehouse_id)
);

CREATE INDEX IF NOT EXISTS idx_product_warehouses_warehouse_id ON product_warehouses(warehouse_id);

-- Seed reference data for development (idempotent)
INSERT INTO categories (id, name, slug) VALUES
    ('11111111-1111-1111-1111-111111111101', 'Engine Parts', 'engine-parts'),
    ('11111111-1111-1111-1111-111111111102', 'Braking System', 'braking-system'),
    ('11111111-1111-1111-1111-111111111103', 'Electrical', 'electrical')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO subcategories (id, category_id, name, slug) VALUES
    ('22222222-2222-2222-2222-222222222201', '11111111-1111-1111-1111-111111111101', 'Filters', 'filters'),
    ('22222222-2222-2222-2222-222222222202', '11111111-1111-1111-1111-111111111102', 'Brake Pads', 'brake-pads')
ON CONFLICT (category_id, slug) DO NOTHING;

INSERT INTO brands (id, name, slug) VALUES
    ('33333333-3333-3333-3333-333333333301', 'Bosch', 'bosch'),
    ('33333333-3333-3333-3333-333333333302', 'Brembo', 'brembo'),
    ('33333333-3333-3333-3333-333333333303', 'Denso', 'denso')
ON CONFLICT (slug) DO NOTHING;
