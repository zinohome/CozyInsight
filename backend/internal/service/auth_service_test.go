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
	"cozy-insight/pkg/jwt"
)

func TestAuthService_Register_Success(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	jwtManager := jwt.NewManager("secret", 2*time.Hour)
	svc := NewAuthService(repo, jwtManager)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE username = (.+) AND deleted_at IS NULL").
		WithArgs("newuser").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec("INSERT INTO users").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "new@example.com",
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Register_Duplicate(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	jwtManager := jwt.NewManager("secret", 2*time.Hour)
	svc := NewAuthService(repo, jwtManager)

	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT (.+) FROM users WHERE username = (.+) AND deleted_at IS NULL").
		WithArgs("existing").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "existing", "hash", "e@x.com", "", "", "", 1, 0, nil, time.Now(), time.Now(), nil,
		))

	err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username: "existing",
		Password: "password123",
		Email:    "new@example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAuthService_Register_CreateError(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	jwtManager := jwt.NewManager("secret", 2*time.Hour)
	svc := NewAuthService(repo, jwtManager)

	// 1. FindByUsername returns no rows
	mock.ExpectQuery("SELECT (.+) FROM users WHERE username = (.+) AND deleted_at IS NULL").
		WithArgs("newuser").
		WillReturnError(sql.ErrNoRows)

	// 2. Insert fails
	mock.ExpectExec("INSERT INTO users").
		WillReturnError(sql.ErrConnDone)

	err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "n@x.com",
		NickName: "New",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create user failed")
}

func TestAuthService_Login_Success(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	jwtManager := jwt.NewManager("secret", 2*time.Hour)
	svc := NewAuthService(repo, jwtManager)

	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at", "deleted_at"}
	// bcrypt hash of "123456" with default cost
	passwd := "123456"
	fromBcrypt, _ := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	mock.ExpectQuery("SELECT (.+) FROM users WHERE username = (.+) AND deleted_at IS NULL").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "alice", string(fromBcrypt), "a@b.com", "", "", "", 1, 0, nil, time.Now(), time.Now(), nil,
		))

	mock.ExpectExec("UPDATE users SET last_login_at").
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "alice",
		Password: passwd,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), resp.UserID)
	assert.Equal(t, "alice", resp.Username)
	assert.NotEmpty(t, resp.Token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	jwtManager := jwt.NewManager("secret", 2*time.Hour)
	svc := NewAuthService(repo, jwtManager)

	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT (.+) FROM users WHERE username = (.+) AND deleted_at IS NULL").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "alice", "wronghash", "a@b.com", "", "", "", 1, 0, nil, time.Now(), time.Now(), nil,
		))

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "alice",
		Password: "wrong",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	jwtManager := jwt.NewManager("secret", 2*time.Hour)
	svc := NewAuthService(repo, jwtManager)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE username = (.+) AND deleted_at IS NULL").
		WithArgs("nobody").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "nobody",
		Password: "123456",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAuthService_Login_UserDisabled(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewUserRepository(db)
	jwtManager := jwt.NewManager("secret", 2*time.Hour)
	svc := NewAuthService(repo, jwtManager)

	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT (.+) FROM users WHERE username = (.+) AND deleted_at IS NULL").
		WithArgs("disabled").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "disabled", "hash", "a@b.com", "", "", "", 0, 0, nil, time.Now(), time.Now(), nil,
		))

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "disabled",
		Password: "123456",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}
