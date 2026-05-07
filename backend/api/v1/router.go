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

	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}
}
