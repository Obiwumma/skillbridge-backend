// Package middleware handles HTTP request preprocessing.
package middleware

import (
	"fmt"
	"net/http"
	"skillbridge-backend/internal/cache"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter tracks and blocks excessive requests from individual client IPs using Redis window buckets.
func RateLimiter(limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cache.RedisClient == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		now := time.Now().Unix()
		bucket := now / int64(window.Seconds())
		key := fmt.Sprintf("ratelimit:%s:%d", ip, bucket)

		ctx := c.Request.Context()
		pipe := cache.RedisClient.TxPipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window*2)

		_, err := pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		count := incr.Val()
		if count > limit {
			response.Error(c, errors.NewAppError(
				http.StatusTooManyRequests,
				"TOO_MANY_REQUESTS",
				"API rate limit exceeded. Please try again later.",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
