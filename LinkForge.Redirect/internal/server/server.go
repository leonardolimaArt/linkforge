package server

import (
	"github.com/gin-gonic/gin"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/handler"
)

func New(
	redirectHandler *handler.RedirectHandler,
	healthHandler *handler.HealthHandler,
) *gin.Engine {
	r := gin.New()

	r.GET("/r/:shortCode", redirectHandler.Handler)
	r.GET("/health", healthHandler.Live)
	r.GET("/ready", healthHandler.Ready)

	return r
}
