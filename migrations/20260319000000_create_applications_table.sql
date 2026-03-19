-- +goose Up
CREATE TYPE application_type AS ENUM ('box', 'special_project');
CREATE TYPE application_source AS ENUM ('telegram_bot', 'manual');
CREATE TYPE application_status AS ENUM ('queue', 'in_progress', 'done');

CREATE TABLE applications (
    id BIGSERIAL PRIMARY KEY,
    type application_type NOT NULL,
    source application_source NOT NULL,
    status application_status NOT NULL DEFAULT 'queue',
    customer_name TEXT NOT NULL,
    contact_info TEXT NOT NULL,
    project_name TEXT,
    box_id BIGINT,
    special_project_id BIGINT,
    manager_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_applications_status ON applications (status);
CREATE INDEX idx_applications_type ON applications (type);
CREATE INDEX idx_applications_manager_id ON applications (manager_id);
CREATE INDEX idx_applications_created_at ON applications (created_at);

-- +goose Down
DROP TABLE IF EXISTS applications;
DROP TYPE IF EXISTS application_status;
DROP TYPE IF EXISTS application_source;
DROP TYPE IF EXISTS application_type;
