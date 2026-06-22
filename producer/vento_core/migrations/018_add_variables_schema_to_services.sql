-- Migration: 018_add_variables_schema_to_services.sql

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'services' AND column_name = 'variables_schema'
    ) THEN
        ALTER TABLE services ADD COLUMN variables_schema JSONB NOT NULL DEFAULT '[]'::jsonb;
    END IF;
END $$;
