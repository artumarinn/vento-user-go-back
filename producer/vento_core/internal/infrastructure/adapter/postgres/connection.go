package postgres

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id         VARCHAR(36)  PRIMARY KEY,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   TEXT         NOT NULL,
    full_name  VARCHAR(255) NOT NULL,
    business_name VARCHAR(255) NOT NULL DEFAULT '',
    google_id  VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='business_name') THEN
        ALTER TABLE users ADD COLUMN business_name VARCHAR(255) NOT NULL DEFAULT '';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='google_id') THEN
        ALTER TABLE users ADD COLUMN google_id VARCHAR(255) UNIQUE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
`

const createProductsTable = `
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS products (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      VARCHAR(36)  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT         NOT NULL,
    sku          TEXT,
    category     TEXT,
    stock        DECIMAL      DEFAULT 0,
    stock_unit   TEXT         DEFAULT 'u',
    max_stock    DECIMAL,
    price        DECIMAL,
    supplier     TEXT,
    supplier_cost DECIMAL,
    metadata     JSONB,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_user_id ON products(user_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_products_updated_at'
    ) THEN
        CREATE TRIGGER update_products_updated_at
        BEFORE UPDATE ON products
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='description') THEN
        ALTER TABLE products ADD COLUMN description TEXT DEFAULT '';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='tags') THEN
        ALTER TABLE products ADD COLUMN tags TEXT DEFAULT '';
    END IF;
END $$;
`

const createBusinessProfilesTable = `
CREATE TABLE IF NOT EXISTS business_profiles (
    id            SERIAL       PRIMARY KEY,
    user_id       VARCHAR(36)  NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    business_name TEXT         NOT NULL DEFAULT '',
    description   TEXT         NOT NULL DEFAULT '',
    industry      TEXT         NOT NULL DEFAULT '',
    tone          TEXT         NOT NULL DEFAULT 'profesional',
    currency      TEXT         NOT NULL DEFAULT 'ARS',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_business_profiles_user_id ON business_profiles(user_id);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='business_profiles' AND column_name='business_name') THEN
        ALTER TABLE business_profiles ADD COLUMN business_name TEXT NOT NULL DEFAULT '';
    END IF;
END $$;
`

const createPaymentsTable = `
CREATE TABLE IF NOT EXISTS payments (
    id              VARCHAR(36)    PRIMARY KEY,
    user_id         VARCHAR(255)   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_name     VARCHAR(255)   NOT NULL,
    amount          DECIMAL(15, 2) NOT NULL,
    status          VARCHAR(50)    NOT NULL DEFAULT 'pendiente_comprobante',
    channel         VARCHAR(100),
    concept         TEXT,
    comprobante_url TEXT,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_status   ON payments(status);
`

const createMetaConfigsTable = `
CREATE TABLE IF NOT EXISTS meta_configs (
    id                       SERIAL       PRIMARY KEY,
    user_id                  VARCHAR(36)  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform_id              VARCHAR(100),
    channel                  VARCHAR(20),
    whatsapp_phone_number_id VARCHAR(50),
    whatsapp_business_id     VARCHAR(50),
    permanent_access_token   TEXT,
    verify_token             TEXT,
    app_secret               TEXT,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meta_configs_user_id ON meta_configs(user_id);

DO $$
BEGIN
    -- Ensure legacy column exists for the migration step if table was created with new schema but migration logic expects it
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='meta_configs' AND column_name='whatsapp_phone_number_id') THEN
        ALTER TABLE meta_configs ADD COLUMN whatsapp_phone_number_id VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='meta_configs' AND column_name='platform_id') THEN
        ALTER TABLE meta_configs ADD COLUMN platform_id VARCHAR(100);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='meta_configs' AND column_name='channel') THEN
        ALTER TABLE meta_configs ADD COLUMN channel VARCHAR(20);
    END IF;

    UPDATE meta_configs SET platform_id = whatsapp_phone_number_id WHERE platform_id IS NULL AND whatsapp_phone_number_id IS NOT NULL;
    UPDATE meta_configs SET channel = 'whatsapp' WHERE channel IS NULL;

    ALTER TABLE meta_configs DROP CONSTRAINT IF EXISTS meta_configs_whatsapp_phone_number_id_key;
    DROP INDEX IF EXISTS idx_meta_configs_whatsapp_phone_number_id;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_meta_configs_platform_id ON meta_configs(platform_id);
CREATE INDEX IF NOT EXISTS idx_meta_configs_user_channel ON meta_configs(user_id, channel);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'meta_configs_channel_check') THEN
        ALTER TABLE meta_configs ADD CONSTRAINT meta_configs_channel_check
            CHECK (channel IN ('whatsapp', 'instagram', 'messenger'));
    END IF;
END $$;
`

const createOrdersTable = `
CREATE TABLE IF NOT EXISTS orders (
    id              VARCHAR(36)    PRIMARY KEY,
    user_id         VARCHAR(36)    NOT NULL REFERENCES users(id),
    client_id       VARCHAR(36)    NOT NULL DEFAULT '',
    client_name     VARCHAR(255),
    conversation_id VARCHAR(255),
    status          VARCHAR(50)    NOT NULL,
    total           DECIMAL(15, 2),
    items           JSONB          NOT NULL DEFAULT '[]'::jsonb,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id         ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_conversation_id ON orders(conversation_id);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='items') THEN
        ALTER TABLE orders ADD COLUMN items JSONB NOT NULL DEFAULT '[]'::jsonb;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='client_id') THEN
        ALTER TABLE orders ADD COLUMN client_id VARCHAR(36) NOT NULL DEFAULT '';
    END IF;
END $$;
`

const createSyncJobsTable = `
CREATE TABLE IF NOT EXISTS sync_jobs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       VARCHAR(50) NOT NULL,
    status     VARCHAR(50) NOT NULL,
    progress   INT         DEFAULT 0,
    error_msg  TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sync_jobs_user_id ON sync_jobs(user_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_sync_jobs_updated_at'
    ) THEN
        CREATE TRIGGER update_sync_jobs_updated_at
        BEFORE UPDATE ON sync_jobs
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
`

const createWebhookEventsTable = `
CREATE TABLE IF NOT EXISTS webhook_events (
    id          BIGSERIAL    PRIMARY KEY,
    event_id    VARCHAR(200) NOT NULL UNIQUE,
    source      VARCHAR(20)  NOT NULL,
    received_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhook_events_received_at ON webhook_events(received_at);
CREATE INDEX IF NOT EXISTS idx_webhook_events_source      ON webhook_events(source);
`

const createPasswordResetsTable = `
CREATE TABLE IF NOT EXISTS password_resets (
    email      VARCHAR(255) NOT NULL,
    token      VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (email, token)
);

CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
`

// NewConnection opens a PostgreSQL connection pool and runs all migrations idempotently.
func NewConnection(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	migrations := []struct {
		name string
		sql  string
	}{
		{"users", createUsersTable},
		{"products", createProductsTable},
		{"business_profiles", createBusinessProfilesTable},
		{"payments", createPaymentsTable},
		{"meta_configs", createMetaConfigsTable},
		{"orders", createOrdersTable},
		{"sync_jobs", createSyncJobsTable},
		{"webhook_events", createWebhookEventsTable},
		{"password_resets", createPasswordResetsTable},
	}

	for _, m := range migrations {
		if _, err := db.Exec(m.sql); err != nil {
			return nil, fmt.Errorf("migration %q failed: %w", m.name, err)
		}
	}

	log.Println("✅ PostgreSQL connected and migrations applied")
	return db, nil
}
