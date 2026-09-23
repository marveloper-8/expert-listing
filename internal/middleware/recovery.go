package middleware

import (
	"log"
	"net/http"

	"expertlisting/internal/models"

	"github.com/gin-gonic/gin"
)

func PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVER] %v", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(
					http.StatusInternalServerError,
					"Internal server error",
					"An unexpected internal error occurred. Please try again later.",
				))
			}
		}()
		c.Next()
	}
}
