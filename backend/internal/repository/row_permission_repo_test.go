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

func TestRowPermissionRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRowPermissionRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO row_permissions").
		WillReturnResult(sqlmock.NewResult(8, 1))

	rp := &model.RowPermission{DatasetID: 1, FieldName: "dept", Operator: "=", Value: "sales", UserAttr: "user.dept", Status: 1}
	require.NoError(t, repo.Create(context.Background(), rp))
	assert.Equal(t, uint64(8), rp.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRowPermissionRepository_ListByDataset(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRowPermissionRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, 1, "dept", "=", "sales", "user.dept", 1, now, now).
			AddRow(2, 1, "region", "IN", "cn,us", "user.region", 1, now, now))

	rps, err := repo.ListByDataset(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, rps, 2)
	assert.Equal(t, "dept", rps[0].FieldName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRowPermissionRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRowPermissionRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("DELETE FROM row_permissions WHERE id = \\?").
		WithArgs(uint64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 3))
	assert.NoError(t, mock.ExpectationsWereMet())
}
