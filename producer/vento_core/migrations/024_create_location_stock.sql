-- Migration: 024_create_location_stock.sql
--
-- Moves stock quantities out of products.stock / insumos.stock (kept for now,
-- deprecated, still written for backward compat) into per-location ledgers.
-- Orders and stock/insumo movements become location-aware too.

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

-- Backfill: copy current aggregate stock into location_stock for each
-- product's default location.
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

-- Backfill: point existing orders and movements at the user's default location.
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
