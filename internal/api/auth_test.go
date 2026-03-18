package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func newAuthHandlersForTest(t *testing.T) (*AuthHandlers, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "postgres")

	h := NewAuthHandlers(sqlxDB, AuthConfig{
		JWTSecret:             "test_secret",
		AccessTokenTTLMinutes: 15,
		RefreshTokenTTLHours:  168,
	}, zap.NewNop())

	return h, mock, func() { _ = db.Close() }
}

func TestRegister_OK(t *testing.T) {
	h, mock, cleanup := newAuthHandlersForTest(t)
	defer cleanup()

	// repo.Create -> GetByEmail -> SELECT * FROM auth_users WHERE email = $1 -> no rows
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM auth_users WHERE email = $1`)).
		WithArgs("admin@example.com").
		WillReturnError(sql.ErrNoRows)

	// INSERT auth_users ... RETURNING ...
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`
INSERT INTO auth_users (name, email, password_hash, role, status)
VALUES ($1, $2, $3, $4, 'active')
RETURNING id, name, email, role, status, created_at, updated_at
`)).
		WithArgs("Admin", "admin@example.com", sqlmock.AnyArg(), "admin").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "email", "role", "status", "created_at", "updated_at",
		}).AddRow(int64(1), "Admin", "admin@example.com", "admin", "active", now, now))

	// refreshTokens.Insert -> Exec INSERT refresh_tokens
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3)
`)).
		WithArgs(sqlmock.AnyArg(), int64(1), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := []byte(`{"name":"Admin","email":"admin@example.com","password":"password123","role":"admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp RegisterResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v body=%s", err, w.Body.String())
	}
	if resp.User == nil || resp.User.Email != "admin@example.com" {
		t.Fatalf("unexpected user: %#v", resp.User)
	}
	if resp.Token == "" || resp.RefreshToken == "" {
		t.Fatalf("expected token + refresh_token, got: token=%q refresh=%q", resp.Token, resp.RefreshToken)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRegister_DuplicateEmail_409(t *testing.T) {
	h, mock, cleanup := newAuthHandlersForTest(t)
	defer cleanup()

	// SELECT by email returns a row => repo.Create -> ErrEmailExists => 409
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM auth_users WHERE email = $1`)).
		WithArgs("admin@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "email", "password_hash", "role", "status", "created_at", "updated_at",
		}).AddRow(int64(1), "Admin", "admin@example.com", "hash", "admin", "active", now, now))

	body := []byte(`{"name":"Admin","email":"admin@example.com","password":"password123","role":"admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d, body=%s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"EMAIL_ALREADY_EXISTS"`)) {
		t.Fatalf("expected EMAIL_ALREADY_EXISTS, body=%s", w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRegister_WeakPassword_400(t *testing.T) {
	h, mock, cleanup := newAuthHandlersForTest(t)
	defer cleanup()

	body := []byte(`{"name":"Admin","email":"admin@example.com","password":"123","role":"admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRefresh_OK_Rotates(t *testing.T) {
	h, mock, cleanup := newAuthHandlersForTest(t)
	defer cleanup()

	oldToken := "old_refresh"

	// Rotate() transaction: Begin -> SELECT ... FOR UPDATE -> UPDATE -> INSERT -> Commit
	mock.ExpectBegin()

	exp := time.Now().Add(24 * time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT token, user_id, expires_at, revoked_at FROM refresh_tokens WHERE token = $1 FOR UPDATE`)).
		WithArgs(oldToken).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "expires_at", "revoked_at"}).
			AddRow(oldToken, int64(1), exp, nil))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1`)).
		WithArgs(oldToken).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES ($1, $2, $3)`)).
		WithArgs(sqlmock.AnyArg(), int64(1), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// role query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT role FROM auth_users WHERE id = $1`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("admin"))

	body := []byte(`{"refresh_token":"old_refresh"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v body=%s", err, w.Body.String())
	}
	if resp["token"] == "" {
		t.Fatalf("expected token, body=%s", w.Body.String())
	}
	if resp["refresh_token"] == "" {
		t.Fatalf("expected refresh_token with rotation, body=%s", w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRefresh_NotFound_401(t *testing.T) {
	h, mock, cleanup := newAuthHandlersForTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT token, user_id, expires_at, revoked_at FROM refresh_tokens WHERE token = $1 FOR UPDATE`)).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	body := []byte(`{"refresh_token":"missing"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body=%s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestLogout_OK_AndThenRefresh401(t *testing.T) {
	h, mock, cleanup := newAuthHandlersForTest(t)
	defer cleanup()

	token := "to_revoke"

	// logout -> UPDATE ... WHERE token = $1 AND revoked_at IS NULL
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1 AND revoked_at IS NULL`)).
		WithArgs(token).
		WillReturnResult(sqlmock.NewResult(0, 1))

	reqLogout := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout",
		bytes.NewReader([]byte(`{"refresh_token":"to_revoke"}`)))
	wLogout := httptest.NewRecorder()
	h.Logout(wLogout, reqLogout)

	if wLogout.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", wLogout.Code, wLogout.Body.String())
	}

	mock.ExpectBegin()
	exp := time.Now().Add(24 * time.Hour)
	revokedAt := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT token, user_id, expires_at, revoked_at FROM refresh_tokens WHERE token = $1 FOR UPDATE`)).
		WithArgs(token).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "expires_at", "revoked_at"}).
			AddRow(token, int64(1), exp, revokedAt))
	mock.ExpectRollback()

	reqRefresh := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh",
		bytes.NewReader([]byte(`{"refresh_token":"to_revoke"}`)))
	wRefresh := httptest.NewRecorder()
	h.Refresh(wRefresh, reqRefresh)

	if wRefresh.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body=%s", wRefresh.Code, wRefresh.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
