-- +goose Up
CREATE TABLE IF NOT EXISTS refresh_tokens (
                                              token TEXT PRIMARY KEY,
                                              user_id BIGINT NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;
