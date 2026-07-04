-- Migration: 023_create_locations.sql

CREATE TABLE IF NOT EXISTS locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    address TEXT,
    phone TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_locations_user_id ON locations(user_id);

-- Only one default location per user.
CREATE UNIQUE INDEX IF NOT EXISTS idx_locations_one_default_per_user
    ON locations(user_id) WHERE is_default;

-- Backfill: every existing user gets a default location so stock/orders can
-- be migrated to it in the next migration.
INSERT INTO locations (id, user_id, name, is_default, created_at)
SELECT gen_random_uuid(), u.id, 'Local Central', TRUE, CURRENT_TIMESTAMP
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM locations l WHERE l.user_id = u.id
);
