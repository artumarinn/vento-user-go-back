-- Documentation only — the real migration is the const array entry in
-- internal/infrastructure/adapter/postgres/connection.go ("default_agent_mode").
ALTER TABLE business_profiles ADD COLUMN IF NOT EXISTS default_agent_mode TEXT NOT NULL DEFAULT 'autonomous';
