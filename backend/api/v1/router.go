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

		// Datasource
		dsRepo := repository.NewDatasourceRepository()
		dsSvc := service.NewDatasourceService(dsRepo)
		dsHandler := handler.NewDatasourceHandler(dsSvc)

		dsGroup := authenticated.Group("/datasource")
		{
			dsGroup.POST("", dsHandler.Create)
			dsGroup.PUT("/:id", dsHandler.Update)
			dsGroup.DELETE("/:id", dsHandler.Delete)
			dsGroup.GET("/:id", dsHandler.Get)
			dsGroup.GET("", dsHandler.List)
			// 新增功能路由
			dsGroup.POST("/:id/validate", dsHandler.Validate)
			dsGroup.GET("/:id/tables", dsHandler.GetTables)
			dsGroup.GET("/:id/fields", dsHandler.GetFields)
		}

		// Dataset
		datasetRepo := repository.NewDatasetRepository()
		datasetSvc := service.NewDatasetService(datasetRepo, engine.Client)
		datasetHandler := handler.NewDatasetHandler(datasetSvc)

		datasetGroup := authenticated.Group("/dataset")
		{
			datasetGroup.POST("/group", datasetHandler.CreateGroup)
			datasetGroup.GET("/group", datasetHandler.ListGroups)
			datasetGroup.POST("/table", datasetHandler.CreateTable)
			datasetGroup.GET("/table", datasetHandler.ListTables)
			datasetGroup.POST("/table/:id/preview", datasetHandler.Preview)
		}

		// Chart
		chartRepo := repository.NewChartRepository()
		datasetRepo = repository.NewDatasetRepository()
		chartSvc := service.NewChartService(chartRepo)
		chartDataSvc := service.NewChartDataService(chartRepo, datasetRepo)
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

		dashboardGroup := authenticated.Group("/dashboard")
		{
			dashboardGroup.POST("", dashboardHandler.Create)
			dashboardGroup.PUT("/:id", dashboardHandler.Update)
			dashboardGroup.DELETE("/:id", dashboardHandler.Delete)
			dashboardGroup.GET("/:id", dashboardHandler.Get)
			dashboardGroup.GET("", dashboardHandler.List)
		}
	}
}
