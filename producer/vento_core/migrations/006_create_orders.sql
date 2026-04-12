CREATE TABLE IF NOT EXISTS orders (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL REFERENCES users(id),
    client_name     VARCHAR(255),
    conversation_id VARCHAR(255),
    status          VARCHAR(50) NOT NULL,
    total           DECIMAL(15,2) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id              SERIAL PRIMARY KEY,
    order_id        VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id      VARCHAR(36) NOT NULL REFERENCES products(id),
    name            VARCHAR(255) NOT NULL,
    quantity        INTEGER NOT NULL,
    price           DECIMAL(15,2) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_conversation_id ON orders(conversation_id);
