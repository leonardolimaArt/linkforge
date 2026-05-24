package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/config"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/handler"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/middleware"
)

func New(
	cfg *config.Config,
	logger *slog.Logger,
	redirectHandler *handler.RedirectHandler,
	healthHandler *handler.HealthHandler,
) *gin.Engine {
	r := gin.New()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logger(logger))

	if cfg.CORSAllowedOrigins != "" {
		r.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	}

	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)

	r.GET("/r/:shortCode", rateLimiter.Middleware(), redirectHandler.Handler)
	r.GET("/health", healthHandler.Live)
	r.GET("/ready", healthHandler.Ready)

	return r
}
