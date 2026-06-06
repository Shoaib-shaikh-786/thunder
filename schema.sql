-- Run this in your NeonDB SQL editor

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    phone         TEXT NOT NULL UNIQUE,
    pin_hash      TEXT NOT NULL,
    type          TEXT NOT NULL CHECK (type IN ('wholesaler', 'dealer', 'salesman', 'staff')),
    wholesaler_id TEXT NOT NULL DEFAULT '',
    dealer_id     TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS invite_tokens (
    token         TEXT PRIMARY KEY,
    wholesaler_id TEXT NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    used          BOOLEAN NOT NULL DEFAULT FALSE
);

-- Index for fast session lookup (middleware runs this on every request)
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Seed your first wholesaler (developer creates this manually)
-- Generate a bcrypt hash of the PIN first, then insert:
-- INSERT INTO users (id, phone, pin_hash, type, wholesaler_id)
-- VALUES ('uuid-here', '9876543210', '$2a$10$...bcrypt_hash...', 'wholesaler', 'uuid-here');