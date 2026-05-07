package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/testutil"
)

func TestUserService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	mock.ExpectQuery("SELECT \\* FROM users WHERE username = \\? AND deleted_at IS NULL").
		WithArgs("newuser").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec("INSERT INTO users").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateUserRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "new@example.com",
		Status:   1,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "newuser", result.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserService_GetByID(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "alice", "hash", "a@b.com", "", "", "", 1, 0, nil, now, now, nil,
		))

	user, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserService_List(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	columns := []string{"id", "username", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT id, username, email, nick_name, avatar, phone, status, is_admin, last_login_at, created_at, updated_at FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "alice", "a@b.com", "", "", "", 1, 0, nil, now, now).
			AddRow(2, "bob", "b@b.com", "", "", "", 1, 0, nil, now, now))

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "alice", list[0].Username)
	assert.Equal(t, "bob", list[1].Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserService_Update(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "alice", "hash", "a@b.com", "", "", "", 1, 0, nil, now, now, nil,
		))

	mock.ExpectExec("UPDATE users SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	status := int8(0)
	isAdmin := true
	err := svc.Update(context.Background(), 1, &dto.UpdateUserRequest{
		Email:    "updated@example.com",
		NickName: "Alice Updated",
		Phone:    "13800138000",
		Status:   &status,
		IsAdmin:  &isAdmin,
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserService_Update_NotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	err := svc.Update(context.Background(), 1, &dto.UpdateUserRequest{Email: "test"})
	assert.Error(t, err)
}

func TestUserService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	mock.ExpectExec("UPDATE users SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserService_ChangePassword(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	// bcrypt hash of "oldpass"
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "alice", string(hash), "a@b.com", "", "", "", 1, 0, nil, now, now, nil,
		))

	mock.ExpectExec("UPDATE users SET password_hash").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.ChangePassword(context.Background(), 1, &dto.ChangePasswordRequest{
		OldPassword: "oldpass",
		NewPassword: "newpass123",
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserService_ChangePassword_WrongOld(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "alice", "wronghash", "a@b.com", "", "", "", 1, 0, nil, now, now, nil,
		))

	err := svc.ChangePassword(context.Background(), 1, &dto.ChangePasswordRequest{
		OldPassword: "wrong",
		NewPassword: "newpass123",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect")
}

func TestUserService_ChangePassword_UserNotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewUserService(repo)

	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	err := svc.ChangePassword(context.Background(), 1, &dto.ChangePasswordRequest{
		OldPassword: "old",
		NewPassword: "newpass123",
	})
	assert.Error(t, err)
}
