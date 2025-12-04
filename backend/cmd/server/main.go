package main

import (
	"flag"
	"fmt"
	"os"

	v1 "cozy-insight-backend/api/v1"
	"cozy-insight-backend/internal/engine"
	"cozy-insight-backend/pkg/config"
	"cozy-insight-backend/pkg/database"
	"cozy-insight-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	configFile := flag.String("config", "configs/app.yaml", "path to config file")
	flag.Parse()

	// 1. Load Config
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Init Logger
	logger.InitLogger(cfg.Logger.Level)
	logger.Log.Info("Starting DataEase Backend...", zap.String("env", cfg.Server.Mode))

	// 3. Init Database
	if err := database.InitDB(cfg.Database); err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Log.Info("Database connected successfully")

	logger.Log.Info("Database connected successfully")

	// 4. Init Calcite Engine
	if err := engine.InitCalciteClient(cfg.Calcite); err != nil {
		logger.Log.Warn("Failed to connect to Avatica Server (SQL Engine might be unavailable)", zap.Error(err))
	} else {
		logger.Log.Info("Calcite Engine connected successfully")
	}

	// 5. Setup Router
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "2.0.0",
		})
	})

	// Register API Routes	// 注册路由
	v1.RegisterRoutes(r, cfg.JWT.Secret)

	// 5. Start Server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Log.Info("Server starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Log.Fatal("Server failed to start", zap.Error(err))
	}
}
