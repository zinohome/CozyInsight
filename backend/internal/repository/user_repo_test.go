package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/model"
)

func TestUserRepository_FindByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM users WHERE username = \\? AND deleted_at IS NULL").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "alice", "hashed", "a@x.com", "Alice", "", "", 1, 1, nil, now, now,
		))

	user, err := repo.FindByUsername(context.Background(), "alice")
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, uint64(1), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "username", "password_hash", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(uint64(7)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			7, "bob", "h", "b@x.com", "Bob", "", "", 1, 0, nil, now, now,
		))

	user, err := repo.FindByID(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, "bob", user.Username)
	assert.Equal(t, int8(0), user.IsAdmin)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO users").
		WillReturnResult(sqlmock.NewResult(99, 1))

	u := &model.User{Username: "carol", PasswordHash: "h", Email: "c@x.com", Status: 1, IsAdmin: 0}
	require.NoError(t, repo.Create(context.Background(), u))
	assert.Equal(t, uint64(99), u.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE users SET last_login_at").
		WithArgs(uint64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UpdateLastLogin(context.Background(), 5))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "username", "email", "nick_name", "avatar", "phone", "status", "is_admin", "last_login_at", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT .* FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "alice", "a@x.com", "Alice", "", "", 1, 1, nil, now, now).
			AddRow(2, "bob", "b@x.com", "Bob", "", "", 1, 0, nil, now, now))

	users, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE users SET email").
		WillReturnResult(sqlmock.NewResult(0, 1))

	u := &model.User{ID: 5, Email: "new@x.com", NickName: "New", Status: 1, IsAdmin: 0}
	require.NoError(t, repo.Update(context.Background(), u))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE users SET password_hash").
		WithArgs("newhash", uint64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UpdatePassword(context.Background(), 3, "newhash"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE users SET deleted_at").
		WithArgs(uint64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 2))
	assert.NoError(t, mock.ExpectationsWereMet())
}
