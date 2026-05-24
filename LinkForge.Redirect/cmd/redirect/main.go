package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/cache"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/config"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/consumer"
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

	redisClient, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		logger.Error("failed to create redis client", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	queries := postgres.New(pool)
	repo := postgres.NewPgRepository(queries)
	redisCache := cache.NewRedisCache(redisClient)
	svc := service.NewRedirectService(repo, redisCache)

	redirectHandler := handler.NewRedirectHandler(svc)
	healthHandler := handler.NewHealthHandler(pool, redisClient)

	engine := server.New(cfg, logger, redirectHandler, healthHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: engine,
	}

	kafkaConsumer := consumer.NewKafkaConsumer(
		consumer.Config{
			Brokers: cfg.KafkaBrokers,
			Topic:   cfg.KafkaTopic,
			GroupID: cfg.KafkaGroupID,
		},
		repo,
		redisCache,
		logger,
	)
	defer kafkaConsumer.Close()

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	consumerDone := make(chan struct{})

	go func() {
		defer close(consumerDone)
		if err := kafkaConsumer.Run(consumerCtx); err != nil {
			logger.Error("consumer stopped with error", "erro", err)
		}
	}()

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shuttingdown signal received")

	logger.Info("stopping kafka consumer")
	cancelConsumer()

	select {
	case <-consumerDone:
		logger.Info("kafka consumer stopped gracefully")
	case <-time.After(10 * time.Second):
		logger.Warn("kafka consumer did not sotp within timeout, proceeding")
	}

	logger.Info("shutting down http server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced http shutdown", "error", err)
	}

	logger.Info("server stopped")
}
