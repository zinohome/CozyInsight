package main

import (
	"cozy-insight-backend/configs"
	"cozy-insight-backend/internal/handler"
	"cozy-insight-backend/internal/middleware"
	"cozy-insight-backend/internal/repository"
	"cozy-insight-backend/internal/service"
	"cozy-insight-backend/pkg/cache"
	"cozy-insight-backend/pkg/database"
	"cozy-insight-backend/pkg/logger"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	configs.LoadConfig()

	// 初始化日志
	logger.InitLogger()

	// 初始化数据库
	database.InitDB()

	// 初始化Redis
	cache.InitRedis()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	r := gin.Default()

	// 中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// 初始化Repository
	authRepo := repository.NewAuthRepository()
	datasourceRepo := repository.NewDatasourceRepository()
	datasetRepo := repository.NewDatasetRepository()
	chartRepo := repository.NewChartRepository()
	dashboardRepo := repository.NewDashboardRepository()
	dashboardComponentRepo := repository.NewDashboardComponentRepository()
	roleRepo := repository.NewRoleRepository()
	permissionRepo := repository.NewPermissionRepository()
	shareRepo := repository.NewShareRepository()
	scheduleRepo := repository.NewScheduleRepository()
	operLogRepo := repository.NewOperLogRepository()
	systemSettingRepo := repository.NewSystemSettingRepository()
	calculatedFieldRepo := repository.NewCalculatedFieldRepository()
	rowPermissionRepo := repository.NewRowPermissionRepository()

	// 初始化Service
	authService := service.NewAuthService(authRepo)
	datasourceService := service.NewDatasourceService(datasourceRepo)
	datasetService := service.NewDatasetService(datasetRepo)
	chartService := service.NewChartService(chartRepo)
	chartDataService := service.NewChartDataService(chartRepo, datasetRepo, nil)
	dashboardService := service.NewDashboardService(dashboardRepo, dashboardComponentRepo)
	roleService := service.NewRoleService(roleRepo)
	permissionService := service.NewPermissionService(permissionRepo, roleRepo)
	exportService := service.NewExportService()
	shareService := service.NewShareService(shareRepo)
	scheduleService := service.NewScheduleService(scheduleRepo)
	operLogService := service.NewOperLogService(operLogRepo)
	systemSettingService := service.NewSystemSettingService(systemSettingRepo)
	calculatedFieldService := service.NewCalculatedFieldService(calculatedFieldRepo)
	datasetGroupService := service.NewDatasetGroupService(datasetRepo)
	rowPermissionService := service.NewRowPermissionService(rowPermissionRepo, roleRepo)

	// 启动定时任务调度器
	if err := scheduleService.Start(); err != nil {
		logger.Log.Fatal("Failed to start schedule service")
	}
	defer scheduleService.Stop()

	// 初始化Handler
	authHandler := handler.NewAuthHandler(authService)
	datasourceHandler := handler.NewDatasourceHandler(datasourceService)
	datasetHandler := handler.NewDatasetHandler(datasetService)
	chartHandler := handler.NewChartHandler(chartService, chartDataService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	roleHandler := handler.NewRoleHandler(roleService)
	permissionHandler := handler.NewPermissionHandler(permissionService)
	exportHandler := handler.NewExportHandler(exportService, datasetService, chartDataService)
	shareHandler := handler.NewShareHandler(shareService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	operLogHandler := handler.NewOperLogHandler(operLogService)
	systemSettingHandler := handler.NewSystemSettingHandler(systemSettingService)
	calculatedFieldHandler := handler.NewCalculatedFieldHandler(calculatedFieldService)
	datasetGroupHandler := handler.NewDatasetGroupHandler(datasetGroupService)
	rowPermissionHandler := handler.NewRowPermissionHandler(rowPermissionService)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 认证相关(公开)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware())
		{
			// 数据源
			datasource := authenticated.Group("/datasource")
			{
				datasource.POST("", datasourceHandler.Create)
				datasource.PUT("/:id", datasourceHandler.Update)
				datasource.DELETE("/:id", datasourceHandler.Delete)
				datasource.GET("/:id", datasourceHandler.Get)
				datasource.GET("", datasourceHandler.List)
				datasource.POST("/test", datasourceHandler.TestConnection)
				datasource.GET("/:id/databases", datasourceHandler.GetDatabases)
				datasource.GET("/:id/tables", datasourceHandler.GetTables)
				datasource.GET("/:id/schema", datasourceHandler.GetTableSchema)
			}

			// 数据集
			dataset := authenticated.Group("/dataset")
			{
				// 分组
				dataset.POST("/group", datasetHandler.CreateGroup)
				dataset.GET("/group", datasetHandler.ListGroups)
				
				// 表
				dataset.POST("/table", datasetHandler.CreateTable)
				dataset.PUT("/table/:id", datasetHandler.UpdateTable)
				dataset.DELETE("/table/:id", datasetHandler.DeleteTable)
				dataset.GET("/table/:id", datasetHandler.GetTable)
				dataset.GET("/table", datasetHandler.ListTables)
				
				// 数据预览
				dataset.GET("/:id/preview", datasetHandler.PreviewData)
				dataset.POST("/:id/sync-fields", datasetHandler.SyncFields)
				
				// 导出
				dataset.GET("/:id/export", exportHandler.ExportDataset)
			}

			// 图表
			chart := authenticated.Group("/chart")
			{
				chart.POST("", chartHandler.Create)
				chart.PUT("/:id", chartHandler.Update)
				chart.DELETE("/:id", chartHandler.Delete)
				chart.GET("/:id", chartHandler.Get)
				chart.GET("", chartHandler.List)
				chart.GET("/:id/data", chartHandler.GetData)
				chart.GET("/:id/export", exportHandler.ExportChartData)
			}

			// 仪表板
			dashboard := authenticated.Group("/dashboard")
			{
				dashboard.POST("", dashboardHandler.Create)
				dashboard.PUT("/:id", dashboardHandler.Update)
				dashboard.DELETE("/:id", dashboardHandler.Delete)
				dashboard.GET("/:id", dashboardHandler.Get)
				dashboard.GET("", dashboardHandler.List)
				dashboard.POST("/:id/publish", dashboardHandler.Publish)
				dashboard.POST("/:id/unpublish", dashboardHandler.Unpublish)
				dashboard.POST("/:id/components", dashboardHandler.SaveComponents)
				dashboard.GET("/:id/components", dashboardHandler.GetComponents)
			}

			// 角色管理
			role := authenticated.Group("/role")
			{
				role.POST("", roleHandler.Create)
				role.PUT("/:id", roleHandler.Update)
				role.DELETE("/:id", roleHandler.Delete)
				role.GET("/:id", roleHandler.Get)
				role.GET("", roleHandler.List)
				role.POST("/:id/assign", roleHandler.AssignToUser)
				role.DELETE("/:id/remove", roleHandler.RemoveFromUser)
				role.GET("/user", roleHandler.GetUserRoles)
			}

			// 权限管理
			permission := authenticated.Group("/permission")
			{
				permission.GET("", permissionHandler.List)
				permission.GET("/role/:roleId", permissionHandler.GetRolePermissions)
				permission.POST("/role/:roleId/grant", permissionHandler.GrantToRole)
				permission.POST("/role/:roleId/revoke", permissionHandler.RevokeFromRole)
				permission.POST("/resource/grant", permissionHandler.GrantResourcePermission)
				permission.GET("/resource", permissionHandler.GetResourcePermissions)
				permission.GET("/check", permissionHandler.CheckPermission)
			}

			// 分享管理
			share := authenticated.Group("/share")
			{
				share.POST("", shareHandler.Create)
				share.GET("/:id", shareHandler.Get)
				share.DELETE("/:id", shareHandler.Delete)
				share.GET("", shareHandler.List)
			}

			// 定时任务
			schedule := authenticated.Group("/schedule")
			{
				schedule.POST("", scheduleHandler.Create)
				schedule.PUT("/:id", scheduleHandler.Update)
				schedule.DELETE("/:id", scheduleHandler.Delete)
				schedule.GET("/:id", scheduleHandler.Get)
				schedule.GET("", scheduleHandler.List)
				schedule.POST("/:id/enable", scheduleHandler.Enable)
				schedule.POST("/:id/disable", scheduleHandler.Disable)
				schedule.POST("/:id/execute", scheduleHandler.Execute)
			}

			// 操作日志
			operLog := authenticated.Group("/log")
			{
				operLog.GET("", operLogHandler.List)
				operLog.POST("/clean", operLogHandler.CleanOld)
			}

			// 系统设置
			setting := authenticated.Group("/setting")
			{
				setting.GET("/:key", systemSettingHandler.Get)
				setting.POST("", systemSettingHandler.Set)
				setting.GET("/type/:type", systemSettingHandler.ListByType)
				setting.DELETE("/:key", systemSettingHandler.Delete)
			}

			// 计算字段管理
			calculatedField := authenticated.Group("/dataset/calculated-field")
			{
				calculatedField.POST("", calculatedFieldHandler.Create)
				calculatedField.GET("/table/:tableId", calculatedFieldHandler.List)
				calculatedField.DELETE("/:id", calculatedFieldHandler.Delete)
			}

			// 数据集分组
			datasetGroup := authenticated.Group("/dataset/group")
			{
				datasetGroup.GET("/tree", datasetGroupHandler.GetTree)
				datasetGroup.GET("", datasetGroupHandler.List)
			}

			// 行级权限
			rowPerm := authenticated.Group("/permission/row")
			{
				rowPerm.POST("/dataset/:datasetId", rowPermissionHandler.Create)
				rowPerm.GET("/dataset/:datasetId", rowPermissionHandler.List)
				rowPerm.DELETE("/:id", rowPermissionHandler.Delete)
			}
		}

		// 公开分享访问(无需认证)
		api.GET("/share/validate/:token", shareHandler.Validate)
		api.GET("/dashboard/public/:id", dashboardHandler.GetPublished)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务器
	port := configs.AppConfig.Server.Port
	logger.Log.Info(fmt.Sprintf("Server starting on port %d", port))
	
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		logger.Log.Fatal(fmt.Sprintf("Failed to start server: %v", err))
	}
}
