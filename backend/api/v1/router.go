package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/handler"
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
	"cozy-insight/pkg/config"
	"cozy-insight/pkg/jwt"
)

func Setup(db *sqlx.DB, cfg *config.Config, r *gin.Engine) {
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtManager)
	authHandler := handler.NewAuthHandler(authService)

	dsRepo := repository.NewDatasourceRepository(db)
	dsService := service.NewDatasourceService(dsRepo)
	dsHandler := handler.NewDatasourceHandler(dsService)

	datasetRepo := repository.NewDatasetRepository(db)
	datasetService := service.NewDatasetService(datasetRepo, dsRepo)
	datasetHandler := handler.NewDatasetHandler(datasetService)

	chartRepo := repository.NewChartRepository(db)
	chartService := service.NewChartService(chartRepo)
	chartHandler := handler.NewChartHandler(chartService)

	dashboardRepo := repository.NewDashboardRepository(db)
	dashboardService := service.NewDashboardService(dashboardRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		authd := api.Group("/")
		authd.Use(middleware.JWTAuth(jwtManager))
		{
			authd.GET("/datasource", dsHandler.List)
			authd.POST("/datasource", dsHandler.Create)
			authd.GET("/datasource/:id", dsHandler.Get)
			authd.PUT("/datasource/:id", dsHandler.Update)
			authd.DELETE("/datasource/:id", dsHandler.Delete)
			authd.POST("/datasource/test", dsHandler.TestConnection)

			authd.GET("/dataset", datasetHandler.List)
			authd.POST("/dataset", datasetHandler.Create)
			authd.GET("/dataset/:id", datasetHandler.Get)
			authd.PUT("/dataset/:id", datasetHandler.Update)
			authd.DELETE("/dataset/:id", datasetHandler.Delete)
			authd.POST("/dataset/:id/sync-fields", datasetHandler.SyncFields)
			authd.GET("/dataset/:id/preview", datasetHandler.Preview)

			authd.GET("/chart", chartHandler.List)
			authd.POST("/chart", chartHandler.Create)
			authd.GET("/chart/:id", chartHandler.Get)
			authd.PUT("/chart/:id", chartHandler.Update)
			authd.DELETE("/chart/:id", chartHandler.Delete)

			authd.GET("/dashboard", dashboardHandler.List)
			authd.POST("/dashboard", dashboardHandler.Create)
			authd.GET("/dashboard/:id", dashboardHandler.Get)
			authd.PUT("/dashboard/:id", dashboardHandler.Update)
			authd.DELETE("/dashboard/:id", dashboardHandler.Delete)
			authd.POST("/dashboard/:id/charts", dashboardHandler.AddChart)
			authd.GET("/dashboard/:id/charts", dashboardHandler.GetCharts)
			authd.DELETE("/dashboard/:id/charts/:chartId", dashboardHandler.RemoveChart)
		}
	}
}
