package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vikash-paf/ragzero/internal/repository"
)

func RateLimitMiddleware(cacheRepo repository.CacheRepository, limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get(TenantContextKey)
		if !exists {
			c.Next()
			return
		}

		key := fmt.Sprintf("rate_limit:%v", tenantID)

		count, err := cacheRepo.Incr(c.Request.Context(), key, 60)
		if err != nil {
			fmt.Printf("Rate limit check failed: %v\n", err)
			c.Next()
			return
		}

		if count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"limit": limit,
			})
			return
		}

		c.Next()
	}
}
