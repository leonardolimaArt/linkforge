package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/service"
)

type RedirectResolver interface {
	Resolve(ctx context.Context, shortCode string) (string, error)
}

type RedirectHandler struct {
	resolver RedirectResolver
}

func NewRedirectHandler(resolver RedirectResolver) *RedirectHandler {
	return &RedirectHandler{resolver: resolver}
}

func (h *RedirectHandler) Handler(c *gin.Context) {
	shortCode := c.Param("shortCode")

	url, err := h.resolver.Resolve(c.Request.Context(), shortCode)
	if errors.Is(err, service.ErrNotFound) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Redirect(http.StatusFound, url)
}
