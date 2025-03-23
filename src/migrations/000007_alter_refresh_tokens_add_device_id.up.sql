ALTER TABLE refresh_tokens
ADD COLUMN device_identifier VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_device_identifier ON refresh_tokens(device_identifier);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_device ON refresh_tokens(user_id, device_identifier); 