package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cozy-insight/pkg/config"
	"cozy-insight/pkg/database"
	"cozy-insight/pkg/logger"
)

func main() {
	cfg := config.Load("configs/app.yaml")

	log := logger.New(cfg.Logger)
	defer log.Sync()

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("failed to connect database", zap.Error(err))
	}
	defer db.Close()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLogger(log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	log.Info("server started", zap.Int("port", cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	log.Info("server exited")
}
