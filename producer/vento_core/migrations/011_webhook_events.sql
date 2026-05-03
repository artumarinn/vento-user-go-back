-- 011_webhook_events.sql
-- Idempotency log for inbound webhook events (Meta, Mercado Pago, etc.).
-- Each row represents one logical event we have already processed.
-- The UNIQUE(event_id) constraint is the dedup primitive: a duplicate INSERT
-- will fail and the handler treats that as "already processed, skip".

CREATE TABLE IF NOT EXISTS webhook_events (
    id           BIGSERIAL PRIMARY KEY,
    event_id     VARCHAR(200) NOT NULL UNIQUE,
    source       VARCHAR(20)  NOT NULL,
    received_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhook_events_received_at ON webhook_events(received_at);
CREATE INDEX IF NOT EXISTS idx_webhook_events_source     ON webhook_events(source);
