-- Add to your existing schema.sql and run in NeonDB
CREATE TABLE IF NOT EXISTS products (
    id text primary key,
    wholesaler_id text NOT NULL references users(id)
    ON
    DELETE
        CASCADE,
        NAME text NOT NULL,
        quantity bigint NOT NULL DEFAULT 0,
        category text NOT NULL DEFAULT '',
        unit text,
        -- nullable: kg, g, ml, etc.
        price bigint NOT NULL DEFAULT 0,
        -- in paise (smallest unit)
        description text NOT NULL DEFAULT '',
        images jsonb NOT NULL DEFAULT '[]',
        -- []*domain.Media stored as JSON
        physical_attributes jsonb,
        -- *domain.PhysicalAttributes, nullable
        created_at bigint NOT NULL,
        -- UnixMilli
        updated_at bigint NOT NULL -- UnixMilli
);
-- Fast list by wholesaler (most common query)
CREATE INDEX IF NOT EXISTS idx_products_wholesaler_id
ON products(wholesaler_id);
-- Filter by category within a wholesaler
CREATE INDEX IF NOT EXISTS idx_products_wholesaler_category
ON products(
    wholesaler_id,
    category
);
-- Full text search on name within a wholesaler
CREATE INDEX IF NOT EXISTS idx_products_name
ON products USING gin(to_tsvector('english', NAME));
