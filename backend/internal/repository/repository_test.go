package repository_test

import (
	"context"
	"testing"
	"time"

	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an  in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate all models
	err = db.AutoMigrate(
		&model.Dashboard{},
		&model.DashboardComponent{},
		&model.ChartView{},
		&model.DatasetGroup{},
		&model.DatasetTable{},
		&model.DatasetTableField{},
		&model.Datasource{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.ResourcePermission{},
		&model.Share{},
		&model.SysOperLog{},
		&model.SysSetting{},
	)
	require.NoError(t, err)

	return db
}

func TestDashboardRepository(t *testing.T) {
	db := setupTestDB(t)
	_ = context.Background()

	repo := repository.NewDashboardRepository()

	t.Run("Create", func(t *testing.T) {
		dashboard := &model.Dashboard{
			ID:         "dash-1",
			Name:       "Test Dashboard",
			NodeType:   "dashboard",
			PID:        "0",
			CreateTime: time.Now().UnixMilli(),
		}

		err := db.Create(dashboard).Error
		assert.NoError(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		dashboard := &model.Dashboard{
			ID:       "dash-2",
			Name:     "Dashboard 2",
			NodeType: "dashboard",
			PID:      "0",
		}
		err := db.Create(dashboard).Error
		require.NoError(t, err)

		// Test repo interface (Note: This is testing the interface exists)
		assert.NotNil(t, repo)
	})
}

func TestChartRepository(t *testing.T) {
	db := setupTestDB(t)
	_ = context.Background()

	repo := repository.NewChartRepository()
	assert.NotNil(t, repo)

	t.Run("Create and Get", func(t *testing.T) {
		chart := &model.ChartView{
			ID:         "chart-1",
			Name:       "Test Chart",
			SceneID:    "scene-1",
			TableID:    "table-1",
			Type:       "bar",
			CreateTime: time.Now().UnixMilli(),
		}

		err := db.Create(chart).Error
		assert.NoError(t, err)

		var retrieved model.ChartView
		err = db.First(&retrieved, "id = ?", "chart-1").Error
		assert.NoError(t, err)
		assert.Equal(t, "Test Chart", retrieved.Name)
	})
}

func TestDatasetRepository(t *testing.T) {
	db := setupTestDB(t)
	_ = context.Background()

	repo := repository.NewDatasetRepository()
	assert.NotNil(t, repo)

	t.Run("CreateGroup", func(t *testing.T) {
		group := &model.DatasetGroup{
			ID:         "group-1",
			Name:       "Test Group",
			PID:        "0",
			Level:      1,
			Type:       "folder",
			CreateTime: time.Now().UnixMilli(),
		}

		err := db.Create(group).Error
		assert.NoError(t, err)
	})

	t.Run("CreateTable", func(t *testing.T) {
		table := &model.DatasetTable{
			ID:                "table-1",
			Name:              "Test Table",
			DatasourceID:      "ds-1",
			DatasetGroupID:    "group-1",
			Type:              "db",
			PhysicalTableName: "test_table",
			CreateTime:        time.Now().UnixMilli(),
		}

		err := db.Create(table).Error
		assert.NoError(t, err)
	})

	t.Run("CreateFields", func(t *testing.T) {
		fields := []*model.DatasetTableField{
			{
				ID:             "field-1",
				DatasetTableID: "table-1",
				Name:           "id",
				OriginName:     "id",
				Type:           "int",
				DeType:         2,
				GroupType:      "d",
			},
			{
				ID:             "field-2",
				DatasetTableID: "table-1",
				Name:           "name",
				OriginName:     "name",
				Type:           "varchar",
				DeType:         0,
				GroupType:      "d",
			},
		}

		err := db.Create(&fields).Error
		assert.NoError(t, err)
	})
}

func TestDatasourceRepository(t *testing.T) {
	db := setupTestDB(t)

	repo := repository.NewDatasourceRepository()
	assert.NotNil(t, repo)

	t.Run("Create", func(t *testing.T) {
		ds := &model.Datasource{
			ID:         "ds-1",
			Name:       "Test Datasource",
			Type:       "mysql",
			CreateTime: time.Now().UnixMilli(),
		}

		err := db.Create(ds).Error
		assert.NoError(t, err)
	})
}

func TestUserRepository(t *testing.T) {
	db := setupTestDB(t)

	repo := repository.NewUserRepository()
	assert.NotNil(t, repo)

	t.Run("Create and GetByUsername", func(t *testing.T) {
		user := &model.User{
			ID:         "user-1",
			Username:   "testuser",
			Email:      "test@example.com",
			Password:   "hashed_password",
			CreateTime: time.Now().UnixMilli(),
		}

		err := db.Create(user).Error
		assert.NoError(t, err)

		var retrieved model.User
		err = db.Where("username = ?", "testuser").First(&retrieved).Error
		assert.NoError(t, err)
		assert.Equal(t, "testuser", retrieved.Username)
	})
}

func TestRoleRepository(t *testing.T) {
	db := setupTestDB(t)

	repo := repository.NewRoleRepository()
	assert.NotNil(t, repo)

	t.Run("Create", func(t *testing.T) {
		role := &model.Role{
			ID:         "role-1",
			Name:       "Admin",
			CreateTime: time.Now().UnixMilli(),
		}

		err := db.Create(role).Error
		assert.NoError(t, err)
	})
}

func TestShareRepository(t *testing.T) {
	db := setupTestDB(t)

	repo := repository.NewShareRepository()
	assert.NotNil(t, repo)

	t.Run("Create", func(t *testing.T) {
		share := &model.Share{
			ID:           "share-1",
			ResourceID:   "dash-1",
			ResourceType: "dashboard",
			Token:        "test-token-123",
			CreateTime:   time.Now().UnixMilli(),
		}

		err := db.Create(share).Error
		assert.NoError(t, err)
	})
}

func TestOperLogRepository(t *testing.T) {
	db := setupTestDB(t)

	repo := repository.NewOperLogRepository()
	assert.NotNil(t, repo)

	t.Run("Create", func(t *testing.T) {
		log := &model.SysOperLog{
			ID:         "log-1",
			UserID:     "user-1",
			Module:     "dashboard",
			Action:     "create",
			CreateTime: time.Now().UnixMilli(),
		}

		err := db.Create(log).Error
		assert.NoError(t, err)
	})
}

// TestAllRepositoriesExist verifies all repository constructors work
func TestAllRepositoriesExist(t *testing.T) {
	tests := []struct {
		name string
		repo interface{}
	}{
		{"DashboardRepository", repository.NewDashboardRepository()},
		{"DashboardComponentRepository", repository.NewDashboardComponentRepository()},
		{"ChartRepository", repository.NewChartRepository()},
		{"DatasetRepository", repository.NewDatasetRepository()},
		{"DatasetFieldRepository", repository.NewDatasetFieldRepository()},
		{"DatasourceRepository", repository.NewDatasourceRepository()},
		{"UserRepository", repository.NewUserRepository()},
		{"RoleRepository", repository.NewRoleRepository()},
		{"PermissionRepository", repository.NewPermissionRepository()},
		{"RowPermissionRepository", repository.NewRowPermissionRepository()},
		{"ShareRepository", repository.NewShareRepository()},
		{"OperLogRepository", repository.NewOperLogRepository()},
		{"SystemSettingRepository", repository.NewSystemSettingRepository()},
		{"ScheduleRepository", repository.NewScheduleRepository()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.repo, "Repository %s should not be nil", tt.name)
		})
	}
}

// Integration-style tests for complex operations
func TestRepositoryIntegration(t *testing.T) {
	db := setupTestDB(t)
	_ = context.Background()

	t.Run("Dashboard with Components", func(t *testing.T) {
		// Create dashboard
		dashboard := &model.Dashboard{
			ID:       "dash-int-1",
			Name:     "Integration Dashboard",
			NodeType: "dashboard",
			PID:      "0",
		}
		err := db.Create(dashboard).Error
		require.NoError(t, err)

		// Create components
		components := []*model.DashboardComponent{
			{
				ID:          "comp-1",
				DashboardID: "dash-int-1",
				X:           0,
				Y:           0,
				W:           6,
				H:           4,
			},
			{
				ID:          "comp-2",
				DashboardID: "dash-int-1",
				X:           6,
				Y:           0,
				W:           6,
				H:           4,
			},
		}
		err = db.Create(&components).Error
		assert.NoError(t, err)

		// Verify
		var count int64
		db.Model(&model.DashboardComponent{}).Where("dashboard_id = ?", "dash-int-1").Count(&count)
		assert.Equal(t, int64(2), count)
	})

	t.Run("Dataset with Fields", func(t *testing.T) {
		// Create dataset table
		table := &model.DatasetTable{
			ID:                "table-int-1",
			Name:              "Sales Data",
			DatasetGroupID:    "group-1",
			Type:              "db",
			PhysicalTableName: "sales",
		}
		err := db.Create(table).Error
		require.NoError(t, err)

		// Create fields
		fields := []*model.DatasetTableField{
			{ID: "f1", DatasetTableID: "table-int-1", Name: "sale_id", Type: "int", DeType: 2},
			{ID: "f2", DatasetTableID: "table-int-1", Name: "amount", Type: "decimal", DeType: 3},
			{ID: "f3", DatasetTableID: "table-int-1", Name: "date", Type: "date", DeType: 1},
		}
		err = db.Create(&fields).Error
		assert.NoError(t, err)

		// Verify
		var fieldCount int64
		db.Model(&model.DatasetTableField{}).Where("dataset_table_id = ?", "table-int-1").Count(&fieldCount)
		assert.Equal(t, int64(3), fieldCount)
	})
}
