package middleware

import (
	"net/http"
	"shorturl/util"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(client *redis.Client, windowSeconds, maxRequests int64) gin.HandlerFunc {
	limiter := util.NewSlidingWindowLimiter(client, windowSeconds, maxRequests)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		allowed, err := limiter.Allow(c.Request.Context(), ip)
		if err != nil {
			c.Next()
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
