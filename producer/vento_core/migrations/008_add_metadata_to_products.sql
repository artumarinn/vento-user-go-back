-- Migration: 008_add_metadata_to_products.sql

ALTER TABLE ia_products ADD COLUMN IF NOT EXISTS metadata JSONB;
