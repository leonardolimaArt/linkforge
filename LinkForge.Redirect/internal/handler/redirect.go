package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/service"
)

type RedirectHandler struct {
	svc *service.RedirectService
}

func NewRedirectHandler(svc *service.RedirectService) *RedirectHandler {
	return &RedirectHandler{svc: svc}
}

func (h *RedirectHandler) Handler(c *gin.Context) {
	shortCode := c.Param("shortCode")

	url, err := h.svc.Resolve(c.Request.Context(), shortCode)
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
