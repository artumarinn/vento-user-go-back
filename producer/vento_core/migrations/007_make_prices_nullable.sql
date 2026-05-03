-- Migration: 007_make_prices_nullable.sql

-- 1. Make price nullable in products
ALTER TABLE products ALTER COLUMN price DROP NOT NULL;
ALTER TABLE products ALTER COLUMN sku DROP NOT NULL;

-- 2. Make total nullable in orders
ALTER TABLE orders ALTER COLUMN total DROP NOT NULL;

-- 3. Create sync_jobs table for tracking async ingestion
CREATE TABLE IF NOT EXISTS sync_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'processing', 'partial_success', 'success', 'failed')),
    progress_pct INT DEFAULT 0,
    result JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_sync_jobs_user_id ON sync_jobs(user_id);

-- Trigger for updated_at on sync_jobs
CREATE TRIGGER update_sync_jobs_updated_at
BEFORE UPDATE ON sync_jobs
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
