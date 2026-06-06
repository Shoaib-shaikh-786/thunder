-- Add to your NeonDB after the existing schema.sql and schema_product.sql
-- Add address column to users table (saved default address)
ALTER TABLE
    users
ADD
    column IF NOT EXISTS address jsonb;
CREATE TABLE IF NOT EXISTS orders (
        id text primary key,
        wholesaler_id text NOT NULL references users(id),
        dealer_id text NOT NULL references users(id),
        placed_by_id text NOT NULL references users(id),
        placed_by_type text NOT NULL CHECK (placed_by_type IN ('dealer', 'salesman')),
        status text NOT NULL DEFAULT 'pending' CHECK (
            status IN (
                'pending',
                'accepted',
                'rejected',
                'processing',
                'shipped',
                'completed',
                'cancelled'
            )
        ),
        items jsonb NOT NULL DEFAULT '[]',
        -- []*OrderItem snapshot
        order_value bigint NOT NULL DEFAULT 0,
        -- in paise
        shipping_address jsonb,
        -- domain.Address snapshot
        etd timestamptz,
        -- set by wholesaler on accept
        notes jsonb NOT NULL DEFAULT '[]',
        -- []*Note
        created_at bigint NOT NULL,
        updated_at bigint NOT NULL
    );
-- Fast list by wholesaler (most common query)
    CREATE INDEX IF NOT EXISTS idx_orders_wholesaler_id
    ON orders(wholesaler_id);
-- Filter by dealer
    CREATE INDEX IF NOT EXISTS idx_orders_dealer_id
    ON orders(dealer_id);
-- Filter by status
    CREATE INDEX IF NOT EXISTS idx_orders_status
    ON orders(
        wholesaler_id,
        status
    );
-- Filter by salesman
    CREATE INDEX IF NOT EXISTS idx_orders_placed_by
    ON orders(placed_by_id);
