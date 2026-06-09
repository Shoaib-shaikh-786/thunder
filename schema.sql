-- ============================================================================
-- UNIFIED SCHEMA FOR THUNDER APPLICATION
-- Tenant & Identity | Product Catalog | Transactions | Billing & Field Intel
-- ============================================================================

-- ============================================================================
-- 1. TENANT & IDENTITY SYSTEM
-- ============================================================================

-- Drop and recreate for clean schema
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS market_knowledge CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS product_media CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS units CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;

CREATE TABLE tenants (
    id VARCHAR(50) PRIMARY KEY,
    company_name VARCHAR(150) NOT NULL,
    business_type VARCHAR(30) NOT NULL,
    branding_config JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    metadata JSONB,
    address JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_tenant_phone ON users (tenant_id, phone);

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- ============================================================================
-- 2. PRODUCT CATALOG ENGINE
-- ============================================================================

CREATE TABLE units (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    abbreviation VARCHAR(10) NOT NULL
);

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    sku VARCHAR(50) NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    category VARCHAR(100),
    description TEXT,
    unit_id INT REFERENCES units(id) ON DELETE SET NULL,
    min_order_qty INT NOT NULL DEFAULT 1,
    is_available BOOLEAN DEFAULT true,
    custom_attributes JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_tenant_sku ON products (tenant_id, sku);
CREATE INDEX idx_products_category ON products (category);
CREATE INDEX idx_products_availability ON products (is_available);
CREATE INDEX idx_products_tenant_id ON products (tenant_id);

CREATE TABLE product_media (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 3. TRANSACTIONS & ORDER PIPELINE
-- ============================================================================

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    buyer_id BIGINT NOT NULL REFERENCES users(id),
    agent_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    total_amount NUMERIC(12,2) NOT NULL,
    status VARCHAR(30) DEFAULT 'pending',
    notes JSONB,
    etd TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orders_tenant_status ON orders (tenant_id, status);
CREATE INDEX idx_orders_buyer ON orders (buyer_id);
CREATE INDEX idx_orders_created_at ON orders (created_at);
CREATE INDEX idx_orders_agent ON orders (agent_id);

CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity INT NOT NULL,
    unit_price NUMERIC(10,2) NOT NULL
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);
CREATE INDEX idx_order_items_product_id ON order_items (product_id);

-- ============================================================================
-- 4. FIELD INTEL & BILLING DOCS
-- ============================================================================

CREATE TABLE market_knowledge (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    agent_id BIGINT NOT NULL REFERENCES users(id),
    buyer_id BIGINT NOT NULL REFERENCES users(id),
    intel_type VARCHAR(50) NOT NULL,
    notes TEXT,
    photo_url VARCHAR(255),
    latitude NUMERIC(9,6),
    longitude NUMERIC(9,6),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_market_intel_type ON market_knowledge (tenant_id, intel_type);
CREATE INDEX idx_market_intel_agent ON market_knowledge (agent_id);
CREATE INDEX idx_market_intel_buyer ON market_knowledge (buyer_id);

CREATE TABLE invoices (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    invoice_number VARCHAR(50) NOT NULL,
    tax_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    grand_total NUMERIC(12,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_invoices_lookup ON invoices (tenant_id, invoice_number);
CREATE INDEX idx_invoices_order_id ON invoices (order_id);