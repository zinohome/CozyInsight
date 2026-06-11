package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/engine"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/testutil"
)

type mockConnector struct {
	columns []engine.ColumnInfo
	data    []map[string]interface{}
	// dataFn lets a test return different data on successive Query calls.
	dataFn func() []map[string]interface{}
}

func (m *mockConnector) Connect(configJSON string) error { return nil }
func (m *mockConnector) Close() error                  { return nil }
func (m *mockConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if m.dataFn != nil {
		return m.dataFn(), nil
	}
	return m.data, nil
}
func (m *mockConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]engine.ColumnInfo, error) {
	return m.columns, nil
}

func TestDatasetService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil, nil)

	mock.ExpectExec("INSERT INTO datasets").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateDatasetRequest{
		Name:         "Test Dataset",
		DatasourceID: 1,
		TableName:    "users",
		Type:         "table",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "Test Dataset", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_Create_SQLDataset(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil, nil)

	mock.ExpectExec("INSERT INTO datasets").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateDatasetRequest{
		Name:         "SQL Dataset",
		DatasourceID: 1,
		TableName:    "orders",
		SQL:          "SELECT * FROM orders",
		Type:         "sql",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "SQL Dataset", result.Name)
	assert.Equal(t, "SELECT * FROM orders", result.SQL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_Create_DBError(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil, nil)

	mock.ExpectExec("INSERT INTO datasets").
		WillReturnError(sql.ErrConnDone)

	_, err := svc.Create(context.Background(), &dto.CreateDatasetRequest{
		Name: "X", DatasourceID: 1, Type: "table",
	}, 1)
	assert.Error(t, err)
}

func TestDatasetService_GetByID(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil, nil)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test", 1, "db", "users", "", "table", 0, 1, 1, now, now, nil,
		))

	ds, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), ds.ID)
	assert.Equal(t, "Test", ds.Name)
	assert.Equal(t, "users", ds.TableName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_inferDeType(t *testing.T) {
	svc := NewDatasetService(nil, nil, nil, nil)

	assert.Equal(t, int8(2), svc.inferDeType("BIGINT"))
	assert.Equal(t, int8(0), svc.inferDeType("VARCHAR"))
	assert.Equal(t, int8(1), svc.inferDeType("DATETIME"))
	assert.Equal(t, int8(2), svc.inferDeType("INT"))
	assert.Equal(t, int8(2), svc.inferDeType("FLOAT"))
	assert.Equal(t, int8(2), svc.inferDeType("DOUBLE"))
	assert.Equal(t, int8(2), svc.inferDeType("DECIMAL"))
	assert.Equal(t, int8(1), svc.inferDeType("DATE"))
	assert.Equal(t, int8(1), svc.inferDeType("TIME"))
	assert.Equal(t, int8(0), svc.inferDeType("TEXT"))
	assert.Equal(t, int8(0), svc.inferDeType("CHAR"))
	assert.Equal(t, int8(4), svc.inferDeType("UNKNOWN"))
}

func TestDatasetService_List(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil, nil)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test", 1, "db", "users", "", "table", 0, 1, 1, now, now, nil,
		))

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_Update(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil, nil)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old", 1, "db", "users", "", "table", 0, 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE datasets SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mode := int8(1)
	status := int8(0)
	dsid := uint64(2)
	err := svc.Update(context.Background(), 1, &dto.UpdateDatasetRequest{
		Name:         "Updated",
		DatasourceID: &dsid,
		DatabaseName: "newdb",
		TableName:    "orders",
		Type:         "sql",
		Mode:         &mode,
		Status:       &status,
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil, nil)

	mock.ExpectExec("UPDATE datasets SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_getTableColumns_UsesConnector(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	mockConn := &mockConnector{columns: []engine.ColumnInfo{
		{Name: "id", Type: "INT", Length: 11},
		{Name: "name", Type: "VARCHAR", Length: 255},
	}}
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	cols, err := svc.getTableColumns(context.Background(), &model.Datasource{Type: "mysql", Config: "{}"}, "db", "users")
	require.NoError(t, err)
	assert.Len(t, cols, 2)
	assert.Equal(t, "id", cols[0].Name)
	assert.Equal(t, "INT", cols[0].Type)
}

func TestDatasetService_queryTableData_UsesConnector(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	mockConn := &mockConnector{data: []map[string]interface{}{{ "id": 42, "name": "Alice" }}}
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	data, err := svc.queryTableData(context.Background(), &model.Datasource{Type: "mysql", Config: "{}"}, "db", "users", 10)
	require.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, 42, data[0]["id"])
}

func TestDatasetService_buildRowFilter(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	rowPermRepo := repository.NewRowPermissionRepository(db)
	svc := NewDatasetService(repo, nil, rowPermRepo, nil)

	columns := []string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, 1, "dept_id", "=", "1", "dept_id", 1, now, now,
		))

	conditions, err := svc.buildRowFilter(context.Background(), 1, map[string]string{"dept_id": "5"})
	require.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "dept_id", conditions[0].FieldName)
	assert.Equal(t, "=", conditions[0].Operator)
	assert.Equal(t, "5", conditions[0].Value)
}

func TestDatasetService_getTableColumns_ConnectorError(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	svc.newConnector = func(string) (engine.DatasourceConnector, error) {
		return nil, fmt.Errorf("unsupported type")
	}

	_, err := svc.getTableColumns(context.Background(), &model.Datasource{Type: "oracle", Config: "{}"}, "db", "users")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create connector failed")
}

func TestDatasetService_queryTableData_ConnectorError(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	svc.newConnector = func(string) (engine.DatasourceConnector, error) {
		return nil, fmt.Errorf("unsupported type")
	}

	_, err := svc.queryTableData(context.Background(), &model.Datasource{Type: "oracle", Config: "{}"}, "db", "users", 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create connector failed")
}

func TestDatasetService_SyncFields_SQL(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "SQL Dataset", 1, "", "", "SELECT id, name FROM orders", "sql", 0, 1, 1, now, now, nil,
		))

	dsColumns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsColumns).AddRow(
			1, "MySQL", "mysql", "{}", 1, 1, now, now, nil,
		))

	mock.ExpectExec("DELETE FROM dataset_fields").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("INSERT INTO dataset_fields").
		WillReturnResult(sqlmock.NewResult(0, 2))

	mockConn := &mockConnector{data: []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}}
	svc.SetConnectorFactory(func(string) (engine.DatasourceConnector, error) { return mockConn, nil })

	err := svc.SyncFields(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_SyncFields_Table(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			2, "Users", 1, "shop", "users", "", "table", 0, 1, 1, now, now, nil,
		))

	dsColumns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsColumns).AddRow(
			1, "MySQL", "mysql", "{}", 1, 1, now, now, nil,
		))

	mock.ExpectExec("DELETE FROM dataset_fields").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("INSERT INTO dataset_fields").
		WillReturnResult(sqlmock.NewResult(0, 3))

	mockConn := &mockConnector{columns: []engine.ColumnInfo{
		{Name: "id", Type: "INT"},
		{Name: "name", Type: "VARCHAR"},
	}}
	svc.SetConnectorFactory(func(string) (engine.DatasourceConnector, error) { return mockConn, nil })

	err := svc.SyncFields(context.Background(), 2)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_SyncFields_DatasetNotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	err := svc.SyncFields(context.Background(), 99)
	assert.Error(t, err)
}

func TestDatasetService_SyncFields_DatasourceNotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "X", 999, "db", "t", "", "table", 0, 1, 1, now, now, nil,
		))

	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	err := svc.SyncFields(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "datasource not found")
}

func TestDatasetService_PreviewData_Table(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	rowPermRepo := repository.NewRowPermissionRepository(db)
	svc := NewDatasetService(repo, dsRepo, rowPermRepo, nil)

	dsColumns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	// 1. FindByID(dataset)
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsColumns).AddRow(
			1, "Users", 5, "shop", "users", "", "table", 0, 1, 1, now, now, nil,
		))
	// 2. FindByID(datasource)
	srcColumns := []string{"id", "name", "type", "config", "file_path", "file_type", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows(srcColumns).AddRow(
			5, "MySQL", "mysql", "{}", "", "", 1, 1, now, now, nil,
		))
	// 3. GetFields
	fColumns := []string{"id", "dataset_id", "name", "type", "de_type", "length", "precision", "scale", "origin_name", "created_at"}
	mock.ExpectQuery("FROM dataset_fields WHERE dataset_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(fColumns).AddRow(1, 1, "id", "INT", 1, 11, 0, 0, "id", now))
	// 4. buildRowFilter: no rows
	mock.ExpectQuery("FROM row_permissions WHERE dataset_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}))

	mockConn := &mockConnector{data: []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}}
	svc.SetConnectorFactory(func(string) (engine.DatasourceConnector, error) { return mockConn, nil })

	resp, err := svc.PreviewData(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Len(t, resp.Fields, 1)
	assert.Equal(t, "id", resp.Fields[0].Name)
	assert.Len(t, resp.Data, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_PreviewData_SQL(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	rowPermRepo := repository.NewRowPermissionRepository(db)
	svc := NewDatasetService(repo, dsRepo, rowPermRepo, nil)

	dsColumns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows(dsColumns).AddRow(
			2, "Sales", 5, "", "", "SELECT * FROM sales", "sql", 0, 1, 1, now, now, nil,
		))
	srcColumns := []string{"id", "name", "type", "config", "file_path", "file_type", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows(srcColumns).AddRow(
			5, "MySQL", "mysql", "{}", "", "", 1, 1, now, now, nil,
		))
	fColumns := []string{"id", "dataset_id", "name", "type", "de_type", "length", "precision", "scale", "origin_name", "created_at"}
	mock.ExpectQuery("FROM dataset_fields WHERE dataset_id = \\?").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows(fColumns))
	mock.ExpectQuery("FROM row_permissions WHERE dataset_id = \\?").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}))

	mockConn := &mockConnector{data: []map[string]interface{}{{"x": 1}}}
	svc.SetConnectorFactory(func(string) (engine.DatasourceConnector, error) { return mockConn, nil })

	resp, err := svc.PreviewData(context.Background(), 2, 5)
	require.NoError(t, err)
	assert.Len(t, resp.Fields, 0)
	assert.Len(t, resp.Data, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_PreviewData_DatasetNotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.PreviewData(context.Background(), 99, 10)
	assert.Error(t, err)
}

func TestDatasetService_getSQLColumns_Happy(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	// First query (LIMIT 0) returns one row of metadata so we don't go down the
	// fallback LIMIT 1 branch.
	mockConn := &mockConnector{data: []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}}
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	cols, err := svc.getSQLColumns(context.Background(), &model.Datasource{Type: "mysql", Config: "{}"}, "SELECT * FROM t")
	require.NoError(t, err)
	assert.Len(t, cols, 2)
	// Both keys should be present; order is map-randomized
	names := []string{cols[0].Name, cols[1].Name}
	assert.Contains(t, names, "id")
	assert.Contains(t, names, "name")
}

func TestDatasetService_getSQLColumns_FallbackLimit1(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	// First query returns 0 rows, second (LIMIT 1) returns 1 row
	calls := 0
	mockConn := &mockConnector{
		dataFn: func() []map[string]interface{} {
			calls++
			if calls == 1 {
				return nil
			}
			return []map[string]interface{}{{"col": 1}}
		},
	}
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	cols, err := svc.getSQLColumns(context.Background(), &model.Datasource{Type: "mysql", Config: "{}"}, "SELECT * FROM t")
	require.NoError(t, err)
	assert.Len(t, cols, 1)
	assert.Equal(t, "col", cols[0].Name)
}

func TestDatasetService_getSQLColumns_NoColumns(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	mockConn := &mockConnector{data: nil} // both queries return nil
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	_, err := svc.getSQLColumns(context.Background(), &model.Datasource{Type: "mysql", Config: "{}"}, "SELECT * FROM empty")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sql returned no columns")
}

func TestDatasetService_getSQLColumns_ConnectorError(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	svc.newConnector = func(string) (engine.DatasourceConnector, error) {
		return nil, fmt.Errorf("nope")
	}

	_, err := svc.getSQLColumns(context.Background(), &model.Datasource{Type: "mysql", Config: "{}"}, "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get connector failed")
}

func TestDatasetService_queryTableDataWithFilter_WithFilter(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	mockConn := &mockConnector{data: []map[string]interface{}{{"id": 1}}}
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	filter := []RowFilterCondition{
		{FieldName: "dept_id", Operator: "=", Value: "5"},
	}
	data, err := svc.queryTableDataWithFilter(
		context.Background(),
		&model.Datasource{Type: "mysql", Config: "{}"},
		"db", "users", 10, filter,
	)
	require.NoError(t, err)
	assert.Len(t, data, 1)
}

func TestDatasetService_queryTableDataWithFilter_NoDbName(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	mockConn := &mockConnector{data: []map[string]interface{}{{"x": 1}}}
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	// No filter, no dbName → tableRef = just "users"
	data, err := svc.queryTableDataWithFilter(
		context.Background(),
		&model.Datasource{Type: "mysql", Config: "{}"},
		"", "users", 5, nil,
	)
	require.NoError(t, err)
	assert.Len(t, data, 1)
}

func TestDatasetService_queryTableDataWithFilter_ConnectorError(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	svc.newConnector = func(string) (engine.DatasourceConnector, error) {
		return nil, fmt.Errorf("boom")
	}

	_, err := svc.queryTableDataWithFilter(
		context.Background(),
		&model.Datasource{Type: "mysql", Config: "{}"},
		"db", "users", 10, nil,
	)
	assert.Error(t, err)
}

func TestDatasetService_querySQLData(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	mockConn := &mockConnector{data: []map[string]interface{}{{"a": 1}, {"a": 2}}}
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	data, err := svc.querySQLData(
		context.Background(),
		&model.Datasource{Type: "mysql", Config: "{}"},
		"SELECT a FROM t", 10,
	)
	require.NoError(t, err)
	assert.Len(t, data, 2)
}

func TestDatasetService_querySQLData_ConnectorError(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil, nil)

	svc.newConnector = func(string) (engine.DatasourceConnector, error) {
		return nil, fmt.Errorf("bad")
	}

	_, err := svc.querySQLData(
		context.Background(),
		&model.Datasource{Type: "mysql", Config: "{}"},
		"SELECT 1", 10,
	)
	assert.Error(t, err)
}
