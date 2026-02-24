-- +goose Up
CREATE TABLE users (
                       id BIGSERIAL PRIMARY KEY,
                       telegram_id BIGINT NOT NULL UNIQUE,
                       username VARCHAR(100),
                       first_name VARCHAR(100),
                       last_name VARCHAR(100),
                       grade SMALLINT DEFAULT 0,  -- 0=external, 1=junior, 2=mid, 3=senior, 4=admin
                       is_admin BOOLEAN DEFAULT FALSE,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_telegram_id ON users(telegram_id);

-- +goose Down
DROP INDEX IF EXISTS idx_users_telegram_id;
DROP TABLE IF EXISTS users;