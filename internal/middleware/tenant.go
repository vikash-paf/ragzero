package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const TenantIDHeader = "X-Tenant-Id"
const TenantContextKey = "tenant_id"

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader(TenantIDHeader)
		if tenantID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-Tenant-Id header"})
			return
		}
		c.Set(TenantContextKey, tenantID)
		c.Next()
	}
}
