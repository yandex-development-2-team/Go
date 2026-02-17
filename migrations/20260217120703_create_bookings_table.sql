-- +goose Up
CREATE TABLE bookings (
                          id BIGSERIAL PRIMARY KEY,
                          user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          service_id SMALLINT NOT NULL,  -- 1=gallery, 2=museum, 3=theater, 4=tennis, 5=padel, 6=digest
                          booking_date DATE NOT NULL,
                          booking_time TIME,  -- опционально для услуг с временем
                          guest_name VARCHAR(255) NOT NULL,
                          guest_organization VARCHAR(255),
                          guest_position VARCHAR(255),
                          visit_type VARCHAR(50),  -- 'private' or 'public'
                          status VARCHAR(50) DEFAULT 'pending',  -- pending, confirmed, cancelled
                          tracker_ticket_id VARCHAR(255),  -- ссылка на Yandex Tracker
                          created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_service_id ON bookings(service_id);
CREATE INDEX idx_bookings_status ON bookings(status);

-- +goose Down
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_service_id;
DROP INDEX IF EXISTS idx_bookings_user_id;
DROP TABLE IF EXISTS bookings;