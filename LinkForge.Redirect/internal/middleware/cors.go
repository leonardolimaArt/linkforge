package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Location"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}
