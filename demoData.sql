-- ============================================================================
-- SEED DEMO DATA FOR THUNDER - FACTORY-A TENANT
-- ============================================================================

-- 1. CREATE TENANT: factory-a
INSERT INTO tenants (id, company_name, business_type, branding_config, is_active)
VALUES (
  'factory-a',
  'Factory A Ltd',
  'factory',
  jsonb_build_object(
    'primary_color', '#FF5733',
    'accent_color', '#33FF57',
    'logo_url', 'https://example.com/factory-a-logo.png'
  ),
  true
) ON CONFLICT (id) DO NOTHING;

-- 2. CREATE UNITS
INSERT INTO units (tenant_id, name, abbreviation)
VALUES 
  ('factory-a', 'Carton Box', 'box'),
  ('factory-a', 'Kilograms', 'kg'),
  ('factory-a', 'Liters', 'L'),
  ('factory-a', 'Pieces', 'pc')
ON CONFLICT DO NOTHING;

-- 3. CREATE DEMO USERS
-- Admin User (PIN: 1234)
INSERT INTO users (tenant_id, name, phone, email, password_hash, role, status, metadata, address)
VALUES (
  'factory-a',
  'Admin User',
  '9876543210',
  'admin@factory-a.com',
  '$2b$10$f3gijDUImQ.GYyqgkEHcEum/ovdLuXW1Txzgmv.aD8xz7b4gfr/F2', -- bcrypt hash of "1234"
  'admin',
  'active',
  jsonb_build_object(
    'department', 'Management',
    'employee_id', 'EMP001'
  ),
  NULL
) ON CONFLICT (tenant_id, phone) DO NOTHING;

-- Field Agent User (PIN: 5678)
INSERT INTO users (tenant_id, name, phone, email, password_hash, role, status, metadata, address)
VALUES (
  'factory-a',
  'Rajesh Kumar',
  '9876543211',
  'rajesh@factory-a.com',
  '$2b$10$g/e1HVSaE7yzZES/en16telI.uf6P1541sV4E5bpF7fBZjI5RF1pa', -- bcrypt hash of "5678"
  'field_agent',
  'active',
  jsonb_build_object(
    'department', 'Field Sales',
    'employee_id', 'EMP002',
    'assigned_region', 'North Karnataka'
  ),
  NULL
) ON CONFLICT (tenant_id, phone) DO NOTHING;

-- Buyer User (PIN: 9999)
INSERT INTO users (tenant_id, name, phone, email, password_hash, role, status, metadata, address)
VALUES (
  'factory-a',
  'Priya Sharma',
  '9876543212',
  'priya@buyer.com',
  '$2b$10$6DnvxeIsmlPXU8NZUbVe2eyVaoqHZ3NOphOInrViW5KMIb2YWAFwu', -- bcrypt hash of "9999"
  'buyer',
  'active',
  jsonb_build_object(
    'shop_name', 'Sharma General Store',
    'business_type', 'retail',
    'credit_limit', 100000
  ),
  jsonb_build_object(
    'street', '123 Market Street',
    'city', 'Bangalore',
    'state', 'Karnataka',
    'postal_code', '560001'
  )
) ON CONFLICT (tenant_id, phone) DO NOTHING;

-- Staff User (PIN: 4321)
INSERT INTO users (tenant_id, name, phone, email, password_hash, role, status, metadata, address)
VALUES (
  'factory-a',
  'Arjun Singh',
  '9876543213',
  'arjun@factory-a.com',
  '$2b$10$vrd7tX5SgaTpJTdKNV2wUu98Eu8npMPf7URq/Sidsr43iUG.iEBbO', -- bcrypt hash of "4321"
  'staff',
  'active',
  jsonb_build_object(
    'department', 'Operations',
    'employee_id', 'EMP003'
  ),
  NULL
) ON CONFLICT (tenant_id, phone) DO NOTHING;

-- 4. CREATE DEMO PRODUCTS
INSERT INTO products (tenant_id, name, sku, price, quantity, category, description, is_available)
VALUES
  ('factory-a', 'Premium Fertilizer', 'FERT-001', 450.00, 1000, 'Fertilizers', 'High-quality NPK fertilizer', true),
  ('factory-a', 'Organic Pesticide', 'PEST-001', 320.00, 500, 'Pesticides', 'Organic certified pesticide', true),
  ('factory-a', 'Agricultural Seeds Pack', 'SEED-001', 150.00, 2000, 'Seeds', 'Mixed vegetable seeds', true),
  ('factory-a', 'Farm Tools Kit', 'TOOL-001', 2500.00, 100, 'Tools', 'Complete farm tools set', true),
  ('factory-a', 'Irrigation Tape', 'IRRI-001', 800.00, 300, 'Irrigation', 'Drip irrigation tape 100m', true)
ON CONFLICT DO NOTHING;