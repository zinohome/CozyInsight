package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/handler"
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

	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		api.GET("/datasource", dsHandler.List)
		api.POST("/datasource", dsHandler.Create)
		api.GET("/datasource/:id", dsHandler.Get)
		api.PUT("/datasource/:id", dsHandler.Update)
		api.DELETE("/datasource/:id", dsHandler.Delete)
		api.POST("/datasource/test", dsHandler.TestConnection)

		api.GET("/dataset", datasetHandler.List)
		api.POST("/dataset", datasetHandler.Create)
		api.GET("/dataset/:id", datasetHandler.Get)
		api.PUT("/dataset/:id", datasetHandler.Update)
		api.DELETE("/dataset/:id", datasetHandler.Delete)
		api.POST("/dataset/:id/sync-fields", datasetHandler.SyncFields)
		api.GET("/dataset/:id/preview", datasetHandler.Preview)
	}
}
