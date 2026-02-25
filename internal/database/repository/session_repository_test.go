package repository

import (
	"context"
	"encoding/json"
	"reflect"
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

	mock.ExpectExec(`INSERT INTO user_sessions`).
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

	mock.ExpectQuery(`FROM user_sessions`).
		WithArgs(int64(10)).
		WillReturnRows(rows)

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

	mock.ExpectExec(`DELETE FROM user_sessions WHERE user_id = \$1`).
		WithArgs(int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

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

	mock.ExpectExec(`UPDATE user_sessions`).
		WithArgs(int64(10), "main_menu").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateSessionState(context.Background(), 10, "main_menu"); err != nil {
		t.Fatalf("UpdateSessionState err: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
