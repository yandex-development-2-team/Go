package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"

	"github.com/yandex-development-2-team/Go/internal/models"
)

// Мок для sql.Result
type mockSQLResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (m *mockSQLResult) LastInsertId() (int64, error) {
	return m.lastInsertId, nil
}

func (m *mockSQLResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

func TestUserRepository_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDatabaseInterface(ctrl)
	logger := zap.NewNop()
	repo := NewUserRepository(mockDB, logger)

	// Успешное создание нового пользователя
	t.Run("success - create new user", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)
		username := "test_user"
		firstName := "Test"
		lastName := "User"
		expectedID := int64(1)

		expectedUser := &models.User{
			ID:         expectedID,
			TelegramID: telegramID,
			Username:   username,
			FirstName:  firstName,
			LastName:   lastName,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			Return(sql.ErrNoRows)
		mockDB.EXPECT().
			ExecContext(
				ctx,
				"INSERT INTO users (telegram_id, username, first_name, last_name) VALUES ($1, $2, $3, $4)",
				telegramID,
				username,
				firstName,
				lastName,
			).
			Return(&mockSQLResult{rowsAffected: 1, lastInsertId: expectedID}, nil)
		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE ID = $1",
				expectedID,
			).
			DoAndReturn(func(ctx context.Context, dest interface{}, query string, id int64) error {
				user := dest.(*models.User)
				*user = *expectedUser
				return nil
			})
		user, err := repo.CreateUser(ctx, telegramID, username, firstName, lastName)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if user == nil {
			t.Fatal("Expected user, got nil")
		}
		if user.ID != expectedID {
			t.Errorf("Expected user ID %d, got %d", expectedID, user.ID)
		}
		if user.TelegramID != telegramID {
			t.Errorf("Expected TelegramID %d, got %d", telegramID, user.TelegramID)
		}
		if user.Username != username {
			t.Errorf("Expected username %s, got %s", username, user.Username)
		}
	})

	// Пользователь уже существует
	t.Run("success - user already exists", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)

		existingUser := &models.User{
			ID:         5,
			TelegramID: telegramID,
			Username:   "existing_user",
			FirstName:  "Existing",
			LastName:   "User",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			DoAndReturn(func(ctx context.Context, dest interface{}, query string, id int64) error {
				user := dest.(*models.User)
				*user = *existingUser
				return nil
			})
		user, err := repo.CreateUser(ctx, telegramID, "new_username", "New", "Name")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if user == nil {
			t.Fatal("Expected user, got nil")
		}
		if user.ID != 5 {
			t.Errorf("Expected existing user with ID 5, got ID %d", user.ID)
		}
		if user.Username != "existing_user" {
			t.Errorf("Expected username 'existing_user', got '%s'", user.Username)
		}
	})

	// Ошибка базы данных
	t.Run("error - database error on insert", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)
		expectedErr := errors.New("insert failed")
		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			Return(sql.ErrNoRows)
		mockDB.EXPECT().
			ExecContext(
				ctx,
				"INSERT INTO users (telegram_id, username, first_name, last_name) VALUES ($1, $2, $3, $4)",
				telegramID,
				"test",
				"Test",
				"User",
			).
			Return(nil, expectedErr)

		_, err := repo.CreateUser(ctx, telegramID, "test", "Test", "User")

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestUserRepository_GetUserByTelegramID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDatabaseInterface(ctrl)
	logger := zap.NewNop()
	repo := NewUserRepository(mockDB, logger)

	// Успешное получение пользователя
	t.Run("success - user found", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)

		expectedUser := &models.User{
			ID:         1,
			TelegramID: telegramID,
			Username:   "test_user",
			FirstName:  "Test",
			LastName:   "User",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			DoAndReturn(func(ctx context.Context, dest interface{}, query string, id int64) error {
				user := dest.(*models.User)
				*user = *expectedUser
				return nil
			})
		user, err := repo.GetUserByTelegramID(ctx, telegramID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.TelegramID, user.TelegramID)
		assert.Equal(t, expectedUser.Username, user.Username)
		assert.Equal(t, expectedUser.FirstName, user.FirstName)
		assert.Equal(t, expectedUser.LastName, user.LastName)
	})

	// Пользователь не найден
	t.Run("error - user not found", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)
		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			Return(sql.ErrNoRows)
		user, err := repo.GetUserByTelegramID(ctx, telegramID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	// Ошибка базы данных
	t.Run("error - database error", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)
		expectedErr := errors.New("connection failed")
		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			Return(expectedErr)
		user, err := repo.GetUserByTelegramID(ctx, telegramID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, expectedErr, err)
	})

	// Контекст отменён до запроса
	t.Run("error - context cancelled before query", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		telegramID := int64(12345)
		user, err := repo.GetUserByTelegramID(ctx, telegramID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, context.Canceled, err)
	})

	// Контекст отменён во время запроса
	t.Run("error - context cancelled during query", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		telegramID := int64(12345)
		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			DoAndReturn(func(ctx context.Context, dest interface{}, query string, id int64) error {
				cancel()
				return ctx.Err()
			})
		user, err := repo.GetUserByTelegramID(ctx, telegramID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.NotNil(t, err)
	})

	// Таймаут контекста
	t.Run("error - context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond)
		telegramID := int64(12345)
		user, err := repo.GetUserByTelegramID(ctx, telegramID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestUserRepository_UpdateUserGrade(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDatabaseInterface(ctrl)
	logger := zap.NewNop()
	repo := NewUserRepository(mockDB, logger)

	// Успешное обновление grade
	t.Run("success - grade updated", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)
		newGrade := 5
		mockDB.EXPECT().
			ExecContext(
				ctx,
				"UPDATE users SET grade = $1 WHERE telegram_id = $2",
				newGrade,
				telegramID,
			).
			Return(&mockSQLResult{rowsAffected: 1}, nil)
		err := repo.UpdateUserGrade(ctx, telegramID, newGrade)
		assert.NoError(t, err)
	})

	// Пользователь не найден (rowsAffected = 0)
	t.Run("error - user not found", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)
		newGrade := 5
		mockDB.EXPECT().
			ExecContext(
				ctx,
				"UPDATE users SET grade = $1 WHERE telegram_id = $2",
				newGrade,
				telegramID,
			).
			Return(&mockSQLResult{rowsAffected: 0}, nil)
		err := repo.UpdateUserGrade(ctx, telegramID, newGrade)
		assert.Error(t, err)
		assert.Equal(t, "no user found", err.Error())
	})

	// Ошибка при выполнении запроса
	t.Run("error - database error", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)
		newGrade := 5
		expectedErr := errors.New("database connection error")
		mockDB.EXPECT().
			ExecContext(
				ctx,
				"UPDATE users SET grade = $1 WHERE telegram_id = $2",
				newGrade,
				telegramID,
			).
			Return(nil, expectedErr)
		err := repo.UpdateUserGrade(ctx, telegramID, newGrade)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	//  Контекст отменён до запроса
	t.Run("error - context cancelled before query", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		telegramID := int64(12345)
		newGrade := 5
		err := repo.UpdateUserGrade(ctx, telegramID, newGrade)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
	// Таймаут контекста
	t.Run("error - context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond)
		telegramID := int64(12345)
		newGrade := 5
		err := repo.UpdateUserGrade(ctx, telegramID, newGrade)
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestUserRepository_IsAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDatabaseInterface(ctrl)
	logger := zap.NewNop()
	repo := NewUserRepository(mockDB, logger)

	// Пользователь является админом
	t.Run("success - user is admin", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)

		expectedUser := &models.User{
			ID:         1,
			TelegramID: telegramID,
			Username:   "admin_user",
			FirstName:  "Admin",
			LastName:   "User",
			IsAdmin:    true,
		}

		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			DoAndReturn(func(ctx context.Context, dest interface{}, query string, id int64) error {
				user := dest.(*models.User)
				*user = *expectedUser
				return nil
			})

		isAdmin, err := repo.IsAdmin(ctx, telegramID)
		assert.NoError(t, err)
		assert.True(t, isAdmin)
	})

	// Пользователь не является админом
	t.Run("success - user is not admin", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)

		expectedUser := &models.User{
			ID:         1,
			TelegramID: telegramID,
			Username:   "regular_user",
			FirstName:  "Regular",
			LastName:   "User",
			IsAdmin:    false,
		}

		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			DoAndReturn(func(ctx context.Context, dest interface{}, query string, id int64) error {
				user := dest.(*models.User)
				*user = *expectedUser
				return nil
			})

		isAdmin, err := repo.IsAdmin(ctx, telegramID)

		assert.NoError(t, err)
		assert.False(t, isAdmin)
	})

	// Пользователь не найден
	t.Run("error - user not found", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)

		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			Return(sql.ErrNoRows)

		isAdmin, err := repo.IsAdmin(ctx, telegramID)

		assert.Error(t, err)
		assert.False(t, isAdmin)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	// Ошибка базы данных
	t.Run("error - database error", func(t *testing.T) {
		ctx := context.Background()
		telegramID := int64(12345)
		expectedErr := errors.New("connection failed")

		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			Return(expectedErr)

		isAdmin, err := repo.IsAdmin(ctx, telegramID)

		assert.Error(t, err)
		assert.False(t, isAdmin)
		assert.Equal(t, expectedErr, err)
	})

	// Контекст отменён до запроса
	t.Run("error - context cancelled before query", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		telegramID := int64(12345)

		isAdmin, err := repo.IsAdmin(ctx, telegramID)

		assert.Error(t, err)
		assert.False(t, isAdmin)
		assert.Equal(t, context.Canceled, err)
	})

	// Контекст отменён во время запроса
	t.Run("error - context cancelled during query", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		telegramID := int64(12345)

		mockDB.EXPECT().
			GetContext(
				ctx,
				gomock.Any(),
				"SELECT * FROM users WHERE telegram_id = $1",
				telegramID,
			).
			DoAndReturn(func(ctx context.Context, dest interface{}, query string, id int64) error {
				cancel() // Отменяем контекст
				return ctx.Err()
			})

		isAdmin, err := repo.IsAdmin(ctx, telegramID)

		assert.Error(t, err)
		assert.False(t, isAdmin)
		assert.Equal(t, context.Canceled, err)
	})

	// Таймаут контекста
	t.Run("error - context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(1 * time.Millisecond)

		telegramID := int64(12345)

		isAdmin, err := repo.IsAdmin(ctx, telegramID)

		assert.Error(t, err)
		assert.False(t, isAdmin)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}
