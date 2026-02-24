package models

import "time"

type UserSession struct {
	ID           int64                  `db:"id"`
	UserID       int64                  `db:"user_id"`
	CurrentState string                 `db:"current_state"`
	StateData    map[string]interface{} `db:"state_data"`
	CreatedAt    time.Time              `db:"created_at"`
	UpdatedAt    time.Time              `db:"updated_at"`
}
