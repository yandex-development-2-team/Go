package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-2-team/Go/internal/models"
	"go.uber.org/zap"
)

func newApplicationRepo(t *testing.T) (*ApplicationRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewApplicationRepository(sqlxDB, zap.NewNop())

	return repo, mock, func() { _ = db.Close() }
}

func TestCreateApplication_OK(t *testing.T) {
	repo, mock, cleanup := newApplicationRepo(t)
	defer cleanup()

	insert := `INSERT INTO applications \(type, source, status, customer_name, contact_info, project_name, box_id, special_project_id\)`
	mock.ExpectQuery(insert).
		WithArgs(models.ApplicationTypeBox, models.ApplicationSourceTelegramBot, models.ApplicationStatusQueue, "Ivan", "+7 999 0000000", nil, nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "source", "status", "customer_name", "contact_info", "project_name", "box_id", "special_project_id", "manager_id", "created_at", "updated_at"}).
			AddRow(int64(1), "box", "telegram_bot", "queue", "Ivan", "+7 999 0000000", nil, nil, nil, nil, time.Now(), time.Now()))

	req := models.ApplicationCreateRequest{
		Type:         models.ApplicationTypeBox,
		Source:       models.ApplicationSourceTelegramBot,
		CustomerName: "Ivan",
		ContactInfo:  "+7 999 0000000",
	}

	app, err := repo.CreateApplication(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateApplication err: %v", err)
	}
	if app == nil || app.ID != 1 {
		t.Fatalf("unexpected application: %#v", app)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestListApplications_WithFilters(t *testing.T) {
	repo, mock, cleanup := newApplicationRepo(t)
	defer cleanup()

	status := models.ApplicationStatusQueue
	typeFilter := models.ApplicationTypeBox
	managerID := int64(10)
	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM applications WHERE 1=1 AND status = \$1 AND type = \$2 AND manager_id = \$3 AND created_at >= \$4 AND created_at <= \$5`).
		WithArgs(status, typeFilter, managerID, from, to).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT id, type, source, status, customer_name, contact_info, project_name, box_id, special_project_id, manager_id, created_at, updated_at FROM applications WHERE 1=1 AND status = \$1 AND type = \$2 AND manager_id = \$3 AND created_at >= \$4 AND created_at <= \$5 ORDER BY created_at DESC LIMIT \$6 OFFSET \$7`).
		WithArgs(status, typeFilter, managerID, from, to, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "source", "status", "customer_name", "contact_info", "project_name", "box_id", "special_project_id", "manager_id", "created_at", "updated_at"}).
			AddRow(int64(1), "box", "telegram_bot", "queue", "Ivan", "+7 999 0000000", nil, nil, nil, managerID, from, to))

	items, total, err := repo.ListApplications(context.Background(), models.ApplicationFilter{
		Status:    &status,
		Type:      &typeFilter,
		ManagerID: &managerID,
		DateFrom:  &from,
		DateTo:    &to,
		Limit:     20,
		Offset:    0,
	})
	if err != nil {
		t.Fatalf("ListApplications err: %v", err)
	}

	if total != 1 {
		t.Fatalf("expected total 1 got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected items 1 got %d", len(items))
	}
	if items[0].CustomerName != "Ivan" {
		t.Fatalf("unexpected item: %#v", items[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

	if !reflect.DeepEqual(items[0].Status, models.ApplicationStatusQueue) {
		t.Fatalf("status mismatch")
	}
}
