package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/cache"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/config"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/handler"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/repository/postgres"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/server"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	defer pool.Close()

	redisClient := cache.NewRedisClient(cfg.RedisURL)
	defer redisClient.Close()

	queries := postgres.New(pool)
	repo := postgres.NewPgRepository(queries)
	redisCache := cache.NewRedisCache(redisClient)
	svc := service.NewRedirectService(repo, redisCache)

	redirectHandler := handler.NewRedirectHandler(svc)
	healthHandler := handler.NewHealthHandler(pool, redisClient)

	engine := server.New(redirectHandler, healthHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: engine,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced shutdown", "error", err)
	}

	logger.Info("server stopped")
}
