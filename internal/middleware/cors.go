package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupCORS(allowOrigins []string) gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if len(allowOrigins) == 1 && allowOrigins[0] == "*" {
		config.AllowAllOrigins = true
		config.AllowOrigins = nil
	}

	return cors.New(config)
}
