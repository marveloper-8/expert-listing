package handlers

import (
	"net/http"
	"time"

	"expertlisting/internal/models"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	startTime time.Time
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// Check godoc
// @Summary Service Health Check
// @Description Returns the service health status, uptime, and timestamp
// @Tags Health
// @Produce json
// @Success 200 {object} models.StandardResponse
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewSuccessResponse(http.StatusOK, "Service is healthy", gin.H{
		"status":    "healthy",
		"uptime":    time.Since(h.startTime).String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}))
}
