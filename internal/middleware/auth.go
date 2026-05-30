package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	TenantIDHeader   = "X-Tenant-Id"
	APIKeyHeader     = "X-API-Key"
	TenantContextKey = "tenant_id"
)

// In a real application, this would be stored in a database, LDAP, or secret manager.
var apiKeyMap = map[string]string{
	"tenant-a": "key-alpha-123",
	"tenant-b": "key-beta-456",
	"tenant-c": "key-gamma-789",
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader(TenantIDHeader)
		apiKey := c.GetHeader(APIKeyHeader)

		if tenantID == "" || apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authentication headers"})
			return
		}

		expectedKey, exists := apiKeyMap[tenantID]
		if !exists || expectedKey != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid tenant id or api key"})
			return
		}

		c.Set(TenantContextKey, tenantID)
		c.Next()
	}
}
