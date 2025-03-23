DROP INDEX IF EXISTS idx_refresh_tokens_device_identifier;
DROP INDEX IF EXISTS idx_refresh_tokens_user_device;

ALTER TABLE refresh_tokens
DROP COLUMN IF EXISTS device_identifier; 