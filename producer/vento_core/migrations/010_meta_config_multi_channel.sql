-- 010_meta_config_multi_channel.sql
-- Allow multiple Meta configurations per user (one per channel: whatsapp, instagram, messenger).
-- Migrates legacy whatsapp_phone_number_id rows into the new platform_id + channel model.

-- 1. Add new columns (nullable first so existing rows survive).
ALTER TABLE meta_configs ADD COLUMN IF NOT EXISTS platform_id VARCHAR(100);
ALTER TABLE meta_configs ADD COLUMN IF NOT EXISTS channel VARCHAR(20);

-- 2. Backfill from legacy column.
UPDATE meta_configs
   SET platform_id = whatsapp_phone_number_id
 WHERE platform_id IS NULL
   AND whatsapp_phone_number_id IS NOT NULL;

UPDATE meta_configs
   SET channel = 'whatsapp'
 WHERE channel IS NULL;

-- 3. Drop the legacy UNIQUE constraint and index on whatsapp_phone_number_id.
ALTER TABLE meta_configs DROP CONSTRAINT IF EXISTS meta_configs_whatsapp_phone_number_id_key;
DROP INDEX IF EXISTS idx_meta_configs_whatsapp_phone_number_id;

-- 4. Enforce NOT NULL on the new columns now that data is backfilled.
ALTER TABLE meta_configs ALTER COLUMN platform_id SET NOT NULL;
ALTER TABLE meta_configs ALTER COLUMN channel SET NOT NULL;

-- 5. New uniqueness + lookup indexes.
CREATE UNIQUE INDEX IF NOT EXISTS idx_meta_configs_platform_id ON meta_configs(platform_id);
CREATE INDEX IF NOT EXISTS idx_meta_configs_user_channel ON meta_configs(user_id, channel);

-- 6. Optional CHECK to keep channel values constrained.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'meta_configs_channel_check'
    ) THEN
        ALTER TABLE meta_configs
            ADD CONSTRAINT meta_configs_channel_check
            CHECK (channel IN ('whatsapp', 'instagram', 'messenger'));
    END IF;
END $$;
