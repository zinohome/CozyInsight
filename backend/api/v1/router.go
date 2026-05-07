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
	rowPermRepo := repository.NewRowPermissionRepository(db)
	datasetService := service.NewDatasetService(datasetRepo, dsRepo, rowPermRepo)
	datasetHandler := handler.NewDatasetHandler(datasetService)
	rowPermService := service.NewRowPermissionService(rowPermRepo)
	rowPermHandler := handler.NewRowPermissionHandler(rowPermService)

	chartRepo := repository.NewChartRepository(db)
	chartService := service.NewChartService(chartRepo, datasetRepo, dsRepo)
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
			authd.GET("/dataset/:id/row-permissions", rowPermHandler.List)
			authd.POST("/dataset/:id/row-permissions", rowPermHandler.Create)
			authd.DELETE("/dataset/:id/row-permissions/:permId", rowPermHandler.Delete)

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

			userRepo := repository.NewUserRepository(db)
			userService := service.NewUserService(userRepo)
			userHandler := handler.NewUserHandler(userService)

			authd.GET("/user/profile", userHandler.Profile)
			authd.POST("/user/change-password", userHandler.ChangePassword)

			admin := authd.Group("/")
			admin.Use(middleware.AdminRequired())
			{
				admin.GET("/user", userHandler.List)
				admin.POST("/user", userHandler.Create)
				admin.GET("/user/:id", userHandler.Get)
				admin.PUT("/user/:id", userHandler.Update)
				admin.DELETE("/user/:id", userHandler.Delete)

				roleRepo := repository.NewRoleRepository(db)
				roleService := service.NewRoleService(roleRepo)
				roleHandler := handler.NewRoleHandler(roleService)

				admin.GET("/role", roleHandler.List)
				admin.POST("/role", roleHandler.Create)
				admin.GET("/role/:id", roleHandler.Get)
				admin.PUT("/role/:id", roleHandler.Update)
				admin.DELETE("/role/:id", roleHandler.Delete)
				admin.GET("/role/menus", roleHandler.ListMenus)
				admin.POST("/role/:id/menus", roleHandler.SetRoleMenus)
				admin.GET("/role/:id/menus", roleHandler.GetRoleMenus)
				admin.POST("/user/:id/roles", roleHandler.SetUserRoles)

				operLogRepo := repository.NewOperationLogRepository(db)
				operLogService := service.NewOperationLogService(operLogRepo)
				operLogHandler := handler.NewOperationLogHandler(operLogService)

				admin.GET("/operation-log", operLogHandler.List)
			}
		}
	}
}
