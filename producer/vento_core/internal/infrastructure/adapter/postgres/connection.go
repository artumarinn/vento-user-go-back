package postgres

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const createUnaccentExtension = `CREATE EXTENSION IF NOT EXISTS unaccent;`

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
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='image_url') THEN
        ALTER TABLE products ADD COLUMN image_url TEXT DEFAULT '';
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_products_user_name_search ON products(user_id, lower(name));
CREATE INDEX IF NOT EXISTS idx_products_user_category    ON products(user_id, category);
`

const createBusinessProfilesTable = `
CREATE TABLE IF NOT EXISTS business_profiles (
    id                 SERIAL       PRIMARY KEY,
    user_id            VARCHAR(36)  NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    business_name      TEXT         NOT NULL DEFAULT '',
    description        TEXT         NOT NULL DEFAULT '',
    industry           TEXT         NOT NULL DEFAULT '',
    tone               TEXT         NOT NULL DEFAULT 'profesional',
    currency           TEXT         NOT NULL DEFAULT 'ARS',
    default_agent_mode TEXT         NOT NULL DEFAULT 'autonomous',
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
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
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='channel') THEN
        ALTER TABLE orders ADD COLUMN channel VARCHAR(20) NOT NULL DEFAULT 'presencial';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='delivery_date') THEN
        ALTER TABLE orders ADD COLUMN delivery_date DATE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='payment_method') THEN
        ALTER TABLE orders ADD COLUMN payment_method VARCHAR(20) NOT NULL DEFAULT 'cash';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='payment_status') THEN
        ALTER TABLE orders ADD COLUMN payment_status VARCHAR(20) NOT NULL DEFAULT 'pending';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='partial_amount') THEN
        ALTER TABLE orders ADD COLUMN partial_amount DECIMAL(15, 2);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='notes') THEN
        ALTER TABLE orders ADD COLUMN notes TEXT;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='payment_recorded_method') THEN
        ALTER TABLE orders ADD COLUMN payment_recorded_method VARCHAR(20);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='payment_recorded_amount') THEN
        ALTER TABLE orders ADD COLUMN payment_recorded_amount DECIMAL(15, 2);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='payment_recorded_at') THEN
        ALTER TABLE orders ADD COLUMN payment_recorded_at TIMESTAMPTZ;
    END IF;
END $$;

UPDATE orders SET status = 'pending' WHERE status = 'seña_pagada';
UPDATE orders SET status = 'processing' WHERE status = 'en_produccion';
UPDATE orders SET status = 'ready' WHERE status = 'listo_entregar';
UPDATE orders SET status = 'delivered' WHERE status = 'entregado';
`

const createOrderStatusHistoryTable = `
CREATE TABLE IF NOT EXISTS order_status_history (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history(order_id);
`

const createOrderPaymentsTable = `
CREATE TABLE IF NOT EXISTS order_payments (
    id         UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   VARCHAR(36)    NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id    VARCHAR(36)    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount     DECIMAL(15, 2) NOT NULL,
    method     VARCHAR(20)    NOT NULL,
    kind       VARCHAR(20)    NOT NULL,
    status     VARCHAR(20)    NOT NULL DEFAULT 'paid',
    paid_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_payments_order_id ON order_payments(order_id);
CREATE INDEX IF NOT EXISTS idx_order_payments_user_id  ON order_payments(user_id);

INSERT INTO order_payments (order_id, user_id, amount, method, kind, status, paid_at, created_at)
SELECT o.id, o.user_id, o.partial_amount, o.payment_method, 'deposit', 'paid', o.created_at, o.created_at
FROM orders o
WHERE o.partial_amount IS NOT NULL
  AND o.partial_amount > 0
  AND NOT EXISTS (
      SELECT 1 FROM order_payments op WHERE op.order_id = o.id AND op.kind = 'deposit'
  );
`

const createClientsTable = `
CREATE TABLE IF NOT EXISTS clients (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='clients' AND column_name='social_network') THEN
        ALTER TABLE clients ADD COLUMN social_network VARCHAR(20);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='clients' AND column_name='social_handle') THEN
        ALTER TABLE clients ADD COLUMN social_handle VARCHAR(255);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_clients_user_id ON clients(user_id);
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

const createInsumosTable = `
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

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_insumos_updated_at'
    ) THEN
        CREATE TRIGGER update_insumos_updated_at
        BEFORE UPDATE ON insumos
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
`

const createTagsTable = `
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#9CA3AF',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, label)
);

CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);

CREATE TABLE IF NOT EXISTS insumo_tags (
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (insumo_id, tag_id)
);
`

const createProductTagsTable = `
CREATE TABLE IF NOT EXISTS product_tags (
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, tag_id)
);
`

const createServiceTagsTable = `
CREATE TABLE IF NOT EXISTS service_tags (
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (service_id, tag_id)
);
`

const createServicesTable = `
CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    formula TEXT NOT NULL,
    minimum_lead_time INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_services_user_id ON services(user_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_services_updated_at'
    ) THEN
        CREATE TRIGGER update_services_updated_at
        BEFORE UPDATE ON services
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'services' AND column_name = 'variables_schema'
    ) THEN
        ALTER TABLE services ADD COLUMN variables_schema JSONB NOT NULL DEFAULT '[]'::jsonb;
    END IF;
END $$;
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

const createStockMovementsTable = `
CREATE TABLE IF NOT EXISTS stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    order_id UUID,
    quantity_delta NUMERIC NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_user_id ON stock_movements(user_id);
`

const createServiceInsumosTable = `
CREATE TABLE IF NOT EXISTS service_insumos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE CASCADE,
    quantity_per_unit NUMERIC NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, insumo_id)
);

CREATE INDEX IF NOT EXISTS idx_service_insumos_service_id ON service_insumos(service_id);
`

const createInsumoMovementsTable = `
CREATE TABLE IF NOT EXISTS insumo_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE CASCADE,
    order_id UUID,
    quantity_delta NUMERIC NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_insumo_movements_insumo_id ON insumo_movements(insumo_id);
`

const createExpensesTable = `
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    concept TEXT NOT NULL,
    category TEXT,
    amount NUMERIC NOT NULL,
    method TEXT,
    paid_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_expenses_user_id ON expenses(user_id);
`

const createLocationsTable = `
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_locations_one_default_per_user
    ON locations(user_id) WHERE is_default;

INSERT INTO locations (id, user_id, name, is_default, created_at)
SELECT gen_random_uuid(), u.id, 'Local Central', TRUE, CURRENT_TIMESTAMP
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM locations l WHERE l.user_id = u.id
);
`

const createLocationStockTables = `
CREATE TABLE IF NOT EXISTS location_stock (
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity DECIMAL NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (location_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_location_stock_product_id ON location_stock(product_id);

CREATE TABLE IF NOT EXISTS location_insumo_stock (
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE CASCADE,
    quantity DECIMAL NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (location_id, insumo_id)
);

CREATE INDEX IF NOT EXISTS idx_location_insumo_stock_insumo_id ON location_insumo_stock(insumo_id);

ALTER TABLE orders ADD COLUMN IF NOT EXISTS location_id UUID REFERENCES locations(id);
ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS location_id UUID REFERENCES locations(id);
ALTER TABLE insumo_movements ADD COLUMN IF NOT EXISTS location_id UUID REFERENCES locations(id);

CREATE INDEX IF NOT EXISTS idx_orders_location_id ON orders(location_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_location_id ON stock_movements(location_id);
CREATE INDEX IF NOT EXISTS idx_insumo_movements_location_id ON insumo_movements(location_id);

INSERT INTO location_stock (location_id, product_id, quantity)
SELECT l.id, p.id, p.stock
FROM products p
JOIN locations l ON l.user_id = p.user_id AND l.is_default = TRUE
ON CONFLICT (location_id, product_id) DO NOTHING;

INSERT INTO location_insumo_stock (location_id, insumo_id, quantity)
SELECT l.id, i.id, COALESCE(i.stock, 0)
FROM insumos i
JOIN locations l ON l.user_id = i.user_id AND l.is_default = TRUE
ON CONFLICT (location_id, insumo_id) DO NOTHING;

UPDATE orders o
SET location_id = l.id
FROM locations l
WHERE l.user_id = o.user_id AND l.is_default = TRUE AND o.location_id IS NULL;

UPDATE stock_movements sm
SET location_id = l.id
FROM locations l
WHERE l.user_id = sm.user_id AND l.is_default = TRUE AND sm.location_id IS NULL;

UPDATE insumo_movements im
SET location_id = l.id
FROM locations l
WHERE l.user_id = im.user_id AND l.is_default = TRUE AND im.location_id IS NULL;
`

const addDefaultAgentModeColumn = `
ALTER TABLE business_profiles ADD COLUMN IF NOT EXISTS default_agent_mode TEXT NOT NULL DEFAULT 'autonomous';
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
		{"unaccent", createUnaccentExtension},
		{"users", createUsersTable},
		{"products", createProductsTable},
		{"business_profiles", createBusinessProfilesTable},
		{"payments", createPaymentsTable},
		{"meta_configs", createMetaConfigsTable},
		{"orders", createOrdersTable},
		{"order_status_history", createOrderStatusHistoryTable},
		{"order_payments", createOrderPaymentsTable},
		{"clients", createClientsTable},
		{"insumos", createInsumosTable},
		{"tags", createTagsTable},
		{"product_tags", createProductTagsTable},
		{"services", createServicesTable},
		{"service_tags", createServiceTagsTable},
		{"sync_jobs", createSyncJobsTable},
		{"webhook_events", createWebhookEventsTable},
		{"password_resets", createPasswordResetsTable},
		{"stock_movements", createStockMovementsTable},
		{"service_insumos", createServiceInsumosTable},
		{"insumo_movements", createInsumoMovementsTable},
		{"expenses", createExpensesTable},
		{"locations", createLocationsTable},
		{"location_stock", createLocationStockTables},
		{"default_agent_mode", addDefaultAgentModeColumn},
	}

	for _, m := range migrations {
		if _, err := db.Exec(m.sql); err != nil {
			// The unaccent extension may require superuser privileges the app's
			// DB role doesn't have on some managed Postgres providers. Product
			// search degrades gracefully to plain ILIKE in that case (see
			// PostgresProductRepository.Search), so don't fail startup over it.
			if m.name == "unaccent" {
				log.Printf("⚠️  could not create unaccent extension (search will fall back to plain ILIKE): %v", err)
				continue
			}
			return nil, fmt.Errorf("migration %q failed: %w", m.name, err)
		}
	}

	log.Println("✅ PostgreSQL connected and migrations applied")
	return db, nil
}
