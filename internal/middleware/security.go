package middleware

import (
	"net/http"
	"sync"
	"time"

	"expertlisting/internal/models"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self' 'unsafe-inline'; img-src 'self' data: https:;")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

func MaxBodySize(limitBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > limitBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, models.NewErrorResponse(
				http.StatusRequestEntityTooLarge,
				"Payload too large",
				"Request body exceeds 1MB limit",
			))
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limitBytes)
		c.Next()
	}
}

type clientLimiter struct {
	tokens     int
	lastRefill time.Time
}

func RateLimiter(maxTokens int, refillRate time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	clients := make(map[string]*clientLimiter)

	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, cl := range clients {
				if now.Sub(cl.lastRefill) > 5*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		cl, exists := clients[ip]
		now := time.Now()

		if !exists {
			clients[ip] = &clientLimiter{
				tokens:     maxTokens - 1,
				lastRefill: now,
			}
			mu.Unlock()
			c.Next()
			return
		}

		elapsed := now.Sub(cl.lastRefill)
		tokensToAdd := int(elapsed / refillRate)
		if tokensToAdd > 0 {
			cl.tokens += tokensToAdd
			if cl.tokens > maxTokens {
				cl.tokens = maxTokens
			}
			cl.lastRefill = now
		}

		if cl.tokens <= 0 {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, models.NewErrorResponse(
				http.StatusTooManyRequests,
				"Rate limit exceeded",
				"Too many requests. Please wait a moment before trying again.",
			))
			return
		}

		cl.tokens--
		mu.Unlock()
		c.Next()
	}
}
