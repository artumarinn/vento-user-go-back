-- 005_create_meta_config.sql
-- Table to store Meta API credentials and configuration for each business

CREATE TABLE IF NOT EXISTS meta_configs (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    whatsapp_phone_number_id VARCHAR(50) UNIQUE,
    whatsapp_business_id VARCHAR(50),
    permanent_access_token TEXT,
    verify_token TEXT,
    app_secret TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_meta_configs_user_id ON meta_configs(user_id);
CREATE INDEX IF NOT EXISTS idx_meta_configs_whatsapp_phone_number_id ON meta_configs(whatsapp_phone_number_id);
