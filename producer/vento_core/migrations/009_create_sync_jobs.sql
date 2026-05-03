-- Migration: 009_create_sync_jobs.sql

CREATE TABLE IF NOT EXISTS sync_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- e.g., 'catalog_ingestion', 'vendor_ingestion'
    status VARCHAR(50) NOT NULL, -- QUEUED, PROCESSING, SUCCESS, FAILED
    progress INT DEFAULT 0,
    error_msg TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sync_jobs_user_id ON sync_jobs(user_id);

CREATE TRIGGER update_sync_jobs_updated_at
BEFORE UPDATE ON sync_jobs
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
