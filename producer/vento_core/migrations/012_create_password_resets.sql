-- 012_create_password_resets.sql
CREATE TABLE IF NOT EXISTS password_resets (
    email      VARCHAR(255) NOT NULL,
    token      VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (email, token)
);

CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
