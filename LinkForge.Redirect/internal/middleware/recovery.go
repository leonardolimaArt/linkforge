package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					"error", r,
					"path", c.Request.URL.Path,
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatus((http.StatusInternalServerError))
			}
		}()
		c.Next()
	}
}
