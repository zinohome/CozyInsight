package v1

import (
	"cozy-insight-backend/internal/engine"
	"cozy-insight-backend/internal/handler"
	"cozy-insight-backend/internal/middleware"
	"cozy-insight-backend/internal/repository"
	"cozy-insight-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, jwtSecret string) {
	api := r.Group("/api/v1")

	// 认证路由 (不需要认证)
	authRepo := repository.NewUserRepository()
	authSvc := service.NewAuthService(authRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authSvc)

	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// 需要认证的路由组
	authenticated := api.Group("")
	authenticated.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// 获取当前用户信息
		authenticated.GET("/auth/me", authHandler.Me)

		// 初始化 Calcite Client（SQL 引擎）
		// TODO: 从配置文件读取这些参数
		calciteClient, err := engine.NewCalciteClient(&engine.CalciteConfig{
			AvaticaURL:      "http://localhost:8765/",
			MaxOpenConns:    100,
			MaxIdleConns:    20,
			ConnMaxLifetime: 1 * 3600 * 1e9, // 1小时（纳秒）
		}, nil) // 暂时不使用缓存
		if err != nil {
			// 在生产环境应该 panic 或返回错误
			// 这里暂时允许继续，但 SQL 功能不可用
			r.GET("/health", func(c *gin.Context) {
				c.JSON(500, gin.H{"error": "Calcite client initialization failed"})
			})
		}

		// 创建数据源连接器
		dsConnector := engine.NewDatasourceConnector(calciteClient)

		// Datasource
		dsRepo := repository.NewDatasourceRepository()
		dsSvc := service.NewDatasourceService(dsRepo, dsConnector)
		dsHandler := handler.NewDatasourceHandler(dsSvc)

		dsGroup := authenticated.Group("/datasource")
		{
			dsGroup.POST("", dsHandler.Create)
			dsGroup.PUT("/:id", dsHandler.Update)
			dsGroup.DELETE("/:id", dsHandler.Delete)
			dsGroup.GET("/:id", dsHandler.Get)
			dsGroup.GET("", dsHandler.List)
			// 连接测试
			dsGroup.POST("/:id/test", dsHandler.TestConnection)
			dsGroup.POST("/test", dsHandler.TestConnectionByConfig)
			// 元数据查询
			dsGroup.GET("/:id/databases", dsHandler.GetDatabases)
			dsGroup.GET("/:id/tables", dsHandler.GetTables)
			dsGroup.GET("/:id/schema", dsHandler.GetTableSchema)
		}

		// Dataset
		datasetRepo := repository.NewDatasetRepository()
		datasetSvc := service.NewDatasetService(datasetRepo, calciteClient)
		datasetHandler := handler.NewDatasetHandler(datasetSvc)

		datasetGroup := authenticated.Group("/dataset")
		{
			datasetGroup.POST("/group", datasetHandler.CreateGroup)
			datasetGroup.GET("/group", datasetHandler.ListGroups)
			datasetGroup.POST("/table", datasetHandler.CreateTable)
			datasetGroup.GET("/table", datasetHandler.ListTables)
			// 数据预览
			datasetGroup.GET("/table/:id/preview", datasetHandler.Preview)
			// 字段管理
			datasetGroup.GET("/table/:id/fields", datasetHandler.GetFields)
			datasetGroup.POST("/table/:id/fields/sync", datasetHandler.SyncFields)
		}

		// Chart
		chartRepo := repository.NewChartRepository()
		datasetRepo = repository.NewDatasetRepository()
		chartSvc := service.NewChartService(chartRepo)
		chartDataSvc := service.NewChartDataService(chartRepo, datasetRepo, calciteClient)
		chartHandler := handler.NewChartHandler(chartSvc, chartDataSvc)

		chartGroup := authenticated.Group("/chart")
		{
			chartGroup.POST("", chartHandler.Create)
			chartGroup.PUT("/:id", chartHandler.Update)
			chartGroup.DELETE("/:id", chartHandler.Delete)
			chartGroup.GET("/:id", chartHandler.Get)
			chartGroup.GET("/:id/data", chartHandler.GetData) // 获取图表数据
			chartGroup.GET("", chartHandler.List)
		}

		// Dashboard
		dashboardRepo := repository.NewDashboardRepository()
		dashboardSvc := service.NewDashboardService(dashboardRepo)
		dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

		dashboards := authenticated.Group("/dashboard")
		{
			dashboards.POST("", dashboardHandler.Create)
			dashboards.PUT("/:id", dashboardHandler.Update)
			dashboards.DELETE("/:id", dashboardHandler.Delete)
			dashboards.GET("/:id", dashboardHandler.Get)
			dashboards.GET("", dashboardHandler.List)

			// 发布相关
			dashboards.POST("/:id/publish", dashboardHandler.Publish)
			dashboards.POST("/:id/unpublish", dashboardHandler.Unpublish)
		}

		// 公开访问（无需认证，但仍在注册函数内）
		r.GET("/api/v1/dashboard/:id/view", dashboardHandler.View)
	}
}
