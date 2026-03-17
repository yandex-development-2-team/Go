-- +goose Up
DO $$ BEGIN
CREATE TYPE user_role_type AS ENUM ('admin', 'manager');
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
CREATE TYPE user_status_type AS ENUM ('active', 'blocked', 'invited');
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS auth_users (
                                          id BIGSERIAL PRIMARY KEY,
                                          name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role user_role_type NOT NULL,
    status user_status_type NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

-- +goose Down
DROP TABLE IF EXISTS auth_users;
DROP TYPE IF EXISTS user_role_type;
DROP TYPE IF EXISTS user_status_type;
