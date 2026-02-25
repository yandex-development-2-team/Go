package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/yandex-development-2-team/Go/internal/models"
)

const (
	saveSessionQuery = `
INSERT INTO user_sessions (user_id, current_state, state_data)
VALUES ($1, $2, $3::jsonb)
ON CONFLICT (user_id) DO UPDATE SET
	current_state = EXCLUDED.current_state,
	state_data = EXCLUDED.state_data,
	updated_at = CURRENT_TIMESTAMP
`
	getSessionQuery = `
SELECT id, user_id, current_state, state_data, created_at, updated_at
FROM user_sessions
WHERE user_id = $1
`
	clearSessionQuery = `DELETE FROM user_sessions WHERE user_id = $1`

	updateSessionStateQuery = `
UPDATE user_sessions
SET current_state = $2,
	updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
`
)

type SessionRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewSessionRepository(db *sqlx.DB, logger *zap.Logger) *SessionRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SessionRepository{db: db, logger: logger}
}

func (r *SessionRepository) SaveSession(ctx context.Context, userID int64, state string, data map[string]interface{}) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}
	if userID <= 0 {
		return fmt.Errorf("invalid userID")
	}
	if data == nil {
		data = map[string]interface{}{}
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal state_data: %w", err)
	}

	_, err = r.db.ExecContext(ctx, saveSessionQuery, userID, state, string(raw))
	if err != nil {
		r.logger.Error("save_session_failed", zap.Error(err), zap.Int64("user_id", userID))
		return fmt.Errorf("save session: %w", err)
	}

	return nil
}

func (r *SessionRepository) GetSession(ctx context.Context, userID int64) (*models.UserSession, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if userID <= 0 {
		return nil, fmt.Errorf("invalid userID")
	}

	var (
		s       models.UserSession
		stateJS []byte
		created time.Time
		updated time.Time
	)

	err := r.db.QueryRowxContext(ctx, getSessionQuery, userID).Scan(
		&s.ID,
		&s.UserID,
		&s.CurrentState,
		&stateJS,
		&created,
		&updated,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get session: %w", err)
	}

	s.CreatedAt = created
	s.UpdatedAt = updated

	if len(stateJS) > 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(stateJS, &m); err != nil {
			return nil, fmt.Errorf("unmarshal state_data: %w", err)
		}
		s.StateData = m
	} else {
		s.StateData = map[string]interface{}{}
	}

	return &s, nil
}

func (r *SessionRepository) ClearSession(ctx context.Context, userID int64) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}
	if userID <= 0 {
		return fmt.Errorf("invalid userID")
	}

	_, err := r.db.ExecContext(ctx, clearSessionQuery, userID)
	if err != nil {
		return fmt.Errorf("clear session: %w", err)
	}
	return nil
}

func (r *SessionRepository) UpdateSessionState(ctx context.Context, userID int64, newState string) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}
	if userID <= 0 {
		return fmt.Errorf("invalid userID")
	}

	_, err := r.db.ExecContext(ctx, updateSessionStateQuery, userID, newState)
	if err != nil {
		return fmt.Errorf("update session state: %w", err)
	}
	return nil
}
