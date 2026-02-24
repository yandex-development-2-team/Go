package repository

import (
	"context"
	"encoding/json"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func newRepo(t *testing.T) (*SessionRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSessionRepository(sqlxDB, zap.NewNop())

	return repo, mock, func() { _ = db.Close() }
}

func TestSaveSession_UpsertJSONB(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	q := regexp.QuoteMeta(`
INSERT INTO user_sessions (user_id, current_state, state_data)
VALUES ($1, $2, $3::jsonb)
ON CONFLICT (user_id) DO UPDATE SET
	current_state = EXCLUDED.current_state,
	state_data = EXCLUDED.state_data,
	updated_at = CURRENT_TIMESTAMP
`)

	mock.ExpectExec(q).
		WithArgs(int64(10), "booking_form", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	payload := map[string]interface{}{
		"step": 3,
		"guest": map[string]interface{}{
			"name": "Ivan",
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal err: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal err: %v", err)
	}

	expected := map[string]interface{}{
		"step": float64(3),
		"guest": map[string]interface{}{
			"name": "Ivan",
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("json mismatch: got=%v expected=%v", got, expected)
	}

	err = repo.SaveSession(context.Background(), 10, "booking_form", payload)
	if err != nil {
		t.Fatalf("SaveSession err: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetSession_OK(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	now := time.Now()

	q := regexp.QuoteMeta(`
SELECT id, user_id, current_state, state_data, created_at, updated_at
FROM user_sessions
WHERE user_id = $1
`)

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "current_state", "state_data", "created_at", "updated_at",
	}).AddRow(
		int64(1),
		int64(10),
		"booking_form",
		[]byte(`{"step":3,"guest":{"name":"Ivan"}}`),
		now,
		now,
	)

	mock.ExpectQuery(q).WithArgs(int64(10)).WillReturnRows(rows)

	s, err := repo.GetSession(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetSession err: %v", err)
	}
	if s == nil {
		t.Fatalf("expected session, got nil")
	}
	if s.CurrentState != "booking_form" {
		t.Fatalf("state mismatch: %s", s.CurrentState)
	}

	guest, ok := s.StateData["guest"].(map[string]interface{})
	if !ok {
		t.Fatalf("guest is not object: %#v", s.StateData["guest"])
	}
	if guest["name"] != "Ivan" {
		t.Fatalf("guest.name mismatch: %#v", guest["name"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestClearSession_OK(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	q := regexp.QuoteMeta(`DELETE FROM user_sessions WHERE user_id = $1`)
	mock.ExpectExec(q).WithArgs(int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.ClearSession(context.Background(), 10); err != nil {
		t.Fatalf("ClearSession err: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpdateSessionState_OK(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	q := regexp.QuoteMeta(`
UPDATE user_sessions
SET current_state = $2,
	updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
`)

	mock.ExpectExec(q).WithArgs(int64(10), "main_menu").WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateSessionState(context.Background(), 10, "main_menu"); err != nil {
		t.Fatalf("UpdateSessionState err: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
