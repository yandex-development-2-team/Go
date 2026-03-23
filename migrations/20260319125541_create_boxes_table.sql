-- +goose Up
CREATE TABLE boxes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    rules TEXT,
    date VARCHAR(8) NOT NULL,
    time VARCHAR(5) NOT NULL, 
    location VARCHAR(255),
    price INTEGER NOT NULL,
    image VARCHAR(500),
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'hidden', 'draft', 'processed')),
    organizer VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT REFERENCES users(id),
    deleted_at TIMESTAMP NULL 
);
CREATE INDEX idx_boxes_status ON boxes(status);
CREATE INDEX idx_boxes_slug ON boxes(slug);
CREATE INDEX idx_boxes_date ON boxes(date);
CREATE INDEX idx_boxes_created_at ON boxes(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_boxes_status;
DROP INDEX IF EXISTS idx_boxes_slug;
DROP INDEX IF EXISTS idx_boxes_date;
DROP INDEX IF EXISTS idx_boxes_created_at;
DROP TABLE IF EXISTS boxes;