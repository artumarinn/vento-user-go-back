-- Migration: 013_create_insumos.sql

CREATE TABLE IF NOT EXISTS insumos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT,
    stock DECIMAL DEFAULT 0,
    stock_unit TEXT DEFAULT 'u',
    price DECIMAL DEFAULT 0,
    supplier TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_insumos_user_id ON insumos(user_id);

CREATE TRIGGER update_insumos_updated_at
BEFORE UPDATE ON insumos
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
