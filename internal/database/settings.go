package database

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func EnsureDefaultSettings(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `
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
`)
	return err
}
