CREATE TABLE refresh_tokens
(
    id                UUID PRIMARY KEY,
    user_id           UUID         NOT NULL REFERENCES users (id),
    token             VARCHAR(255) NOT NULL,
    expires_at        TIMESTAMP    NOT NULL,
    revoked           BOOLEAN   DEFAULT FALSE,
    created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by_ip     VARCHAR(45),
    revoked_at        TIMESTAMP,
    revoked_by_ip     VARCHAR(45),
    replaced_by_token VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens (token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);