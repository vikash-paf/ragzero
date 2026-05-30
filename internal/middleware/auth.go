package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	TenantIDHeader   = "X-Tenant-Id"
	APIKeyHeader     = "X-API-Key"
	TenantContextKey = "tenant_id"
)

var (
	apiKeyMap = make(map[string]string)
	mu        sync.RWMutex
)

// LoadAPIKeys loads keys from a JSON file. In production, this might be a vault call or a DB lookup.
func LoadAPIKeys(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()
	return json.Unmarshal(file, &apiKeyMap)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader(TenantIDHeader)
		apiKey := c.GetHeader(APIKeyHeader)

		if tenantID == "" || apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authentication headers"})
			return
		}

		mu.RLock()
		expectedKey, exists := apiKeyMap[tenantID]
		mu.RUnlock()

		if !exists || expectedKey != apiKey {
			slog.Warn("Unauthorized access attempt", "tenant_id", tenantID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid tenant id or api key"})
			return
		}

		c.Set(TenantContextKey, tenantID)
		c.Next()
	}
}
