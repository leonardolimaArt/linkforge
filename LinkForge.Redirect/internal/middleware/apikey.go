package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKey(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedKey == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		provided := c.GetHeader("X-API-Key")
		if provided != expectedKey {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}
