-- ============================================================================
-- MIGRATION: Add missing address column and ensure schema consistency
-- Run this to fix the existing database if needed
-- ============================================================================

-- Check and add address column if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name='users' AND column_name='address'
    ) THEN
        ALTER TABLE users ADD COLUMN address JSONB;
    END IF;
END $$;

-- Verify all required columns exist
DO $$ 
BEGIN
    -- Check email column
    IF NOT EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name='users' AND column_name='email'
    ) THEN
        ALTER TABLE users ADD COLUMN email VARCHAR(100);
    END IF;
    
    -- Check metadata column
    IF NOT EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name='users' AND column_name='metadata'
    ) THEN
        ALTER TABLE users ADD COLUMN metadata JSONB;
    END IF;
END $$;

-- Ensure indexes exist
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_phone ON users (tenant_id, phone);

-- Verify the schema
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns 
WHERE table_name='users'
ORDER BY ordinal_position;
