package v1

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/engine"
	"cozy-insight/internal/handler"
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
	"cozy-insight/pkg/cache"
	"cozy-insight/pkg/config"
	"cozy-insight/pkg/jwt"
)

func Setup(db *sqlx.DB, cfg *config.Config, r *gin.Engine, redisClient *cache.RedisClient) {
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtManager)
	authHandler := handler.NewAuthHandler(authService)

	dsRepo := repository.NewDatasourceRepository(db)
	connectorPool := engine.NewConnectorPool()
	dsService := service.NewDatasourceService(dsRepo, connectorPool)
	dsHandler := handler.NewDatasourceHandler(dsService)

	datasetRepo := repository.NewDatasetRepository(db)
	rowPermRepo := repository.NewRowPermissionRepository(db)
	datasetService := service.NewDatasetService(datasetRepo, dsRepo, rowPermRepo, connectorPool)
	datasetHandler := handler.NewDatasetHandler(datasetService)
	rowPermService := service.NewRowPermissionService(rowPermRepo)
	rowPermHandler := handler.NewRowPermissionHandler(rowPermService)

	chartRepo := repository.NewChartRepository(db)
	var cacheService *service.CacheService
	if redisClient != nil {
		cacheService = service.NewCacheService(redisClient)
		if err := redisClient.Ping(context.Background()); err != nil {
			fmt.Printf("redis connection failed: %v\n", err)
			cacheService = nil
		}
	}
	chartService := service.NewChartService(chartRepo, datasetRepo, dsRepo, cacheService, connectorPool)
	chartHandler := handler.NewChartHandler(chartService)
	exportHandler := handler.NewExportHandler(chartService)

	dashboardRepo := repository.NewDashboardRepository(db)
	shareLinkRepo := repository.NewShareLinkRepository(db)
	dashboardService := service.NewDashboardService(dashboardRepo, shareLinkRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	shareLinkService := service.NewShareLinkService(shareLinkRepo, dashboardRepo)
	shareHandler := handler.NewShareHandler(shareLinkService)

	workbenchRepo := repository.NewWorkbenchRepository(db)
	workbenchService := service.NewWorkbenchService(workbenchRepo)
	workbenchHandler := handler.NewWorkbenchHandler(workbenchService)

	messageRepo := repository.NewMessageRepository(db)
	messageService := service.NewMessageService(messageRepo)
	messageHandler := handler.NewMessageHandler(messageService)

	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.GET("/share/:token", shareHandler.GetDashboard)

		authd := api.Group("/")
		authd.Use(middleware.JWTAuth(jwtManager))
		{
			authd.GET("/datasource", dsHandler.List)
			authd.POST("/datasource", dsHandler.Create)
			authd.GET("/datasource/:id", dsHandler.Get)
			authd.PUT("/datasource/:id", dsHandler.Update)
			authd.DELETE("/datasource/:id", dsHandler.Delete)
			authd.POST("/datasource/test", dsHandler.TestConnection)
				authd.POST("/datasource/upload", dsHandler.UploadFile)

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
			authd.POST("/chart/:id/data", chartHandler.GetData)
			authd.GET("/chart/:id/export/csv", exportHandler.ExportCSV)
			authd.GET("/chart/:id/export/excel", exportHandler.ExportExcel)

			authd.GET("/dashboard", dashboardHandler.List)
			authd.POST("/dashboard", dashboardHandler.Create)
			authd.GET("/dashboard/:id", dashboardHandler.Get)
			authd.PUT("/dashboard/:id", dashboardHandler.Update)
			authd.DELETE("/dashboard/:id", dashboardHandler.Delete)
			authd.POST("/dashboard/:id/charts", dashboardHandler.AddChart)
			authd.GET("/dashboard/:id/charts", dashboardHandler.GetCharts)
			authd.DELETE("/dashboard/:id/charts/:chartId", dashboardHandler.RemoveChart)
				authd.POST("/dashboard/:id/share", dashboardHandler.EnableShare)
				authd.DELETE("/dashboard/:id/share", dashboardHandler.DisableShare)
				authd.GET("/share-links", shareHandler.List)

				authd.GET("/workbench/stats", workbenchHandler.GetStats)
				authd.GET("/workbench/recent", workbenchHandler.GetRecentViews)
				authd.POST("/workbench/recent", workbenchHandler.RecordVisit)
				authd.GET("/workbench/favorites", workbenchHandler.GetFavorites)
				authd.POST("/workbench/favorites", workbenchHandler.AddFavorite)
				authd.DELETE("/workbench/favorites/:type/:id", workbenchHandler.DeleteFavorite)

				authd.GET("/messages", messageHandler.List)
				authd.GET("/messages/unread-count", messageHandler.CountUnread)
				authd.POST("/messages/:id/read", messageHandler.MarkAsRead)
				authd.POST("/messages/read-all", messageHandler.MarkAllAsRead)
				authd.DELETE("/messages/:id", messageHandler.Delete)

				scheduleRepo := repository.NewScheduleTaskRepository(db)
				scheduleService := service.NewScheduleService(scheduleRepo)
				scheduleHandler := handler.NewScheduleHandler(scheduleService)

				authd.GET("/schedule", scheduleHandler.List)
				authd.POST("/schedule", scheduleHandler.Create)
				authd.GET("/schedule/:id", scheduleHandler.Get)
				authd.PUT("/schedule/:id", scheduleHandler.Update)
				authd.DELETE("/schedule/:id", scheduleHandler.Delete)
				authd.POST("/schedule/:id/enable", scheduleHandler.Enable)
				authd.POST("/schedule/:id/disable", scheduleHandler.Disable)
				authd.POST("/schedule/:id/execute", scheduleHandler.Execute)

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
