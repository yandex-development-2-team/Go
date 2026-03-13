-- +goose Up
CREATE TABLE IF NOT EXISTS settings (
    id SMALLINT PRIMARY KEY,
    notifications JSONB NOT NULL,
    booking JSONB NOT NULL,
    general JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO settings (id, notifications, booking, general)
VALUES (
           1,
           '{
             "telegram_bot_token": "",
             "auto_reminders": false,
             "reminder_hours_before": 24
           }'::jsonb,
           '{
             "max_slots_per_event": 10,
             "allow_overbooking": false,
             "cancellation_allowed_hours": 24
           }'::jsonb,
           '{
             "site_name": "Yandex Bot",
             "contact_email": "support@example.com",
             "contact_phone": ""
           }'::jsonb
       )
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS settings;
