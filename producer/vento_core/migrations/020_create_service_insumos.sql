CREATE TABLE IF NOT EXISTS service_insumos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    insumo_id UUID NOT NULL REFERENCES insumos(id) ON DELETE CASCADE,
    quantity_per_unit NUMERIC NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, insumo_id)
);
CREATE INDEX IF NOT EXISTS idx_service_insumos_service_id ON service_insumos(service_id);
