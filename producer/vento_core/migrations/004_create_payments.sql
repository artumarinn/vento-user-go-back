CREATE TABLE IF NOT EXISTS payments (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_name VARCHAR(255) NOT NULL,
    amount      DECIMAL(15, 2) NOT NULL,
    status      VARCHAR(50) NOT NULL DEFAULT 'pendiente_comprobante',
    channel     VARCHAR(100),
    concept     TEXT,
    comprobante_url TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
