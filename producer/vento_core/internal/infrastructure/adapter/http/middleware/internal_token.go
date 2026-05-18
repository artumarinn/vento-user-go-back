package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// InternalTokenMiddleware protects internal routes with a shared secret.
// Expects X-Internal-Token (shared secret) and X-Tenant-User-Id (target tenant).
func InternalTokenMiddleware(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Internal-Token")
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "unauthorized", "message": "invalid internal token"},
			})
			return
		}

		tenantID := c.GetHeader("X-Tenant-User-Id")
		if tenantID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "bad_request", "message": "X-Tenant-User-Id header is required"},
			})
			return
		}

		c.Set("tenantUserID", tenantID)
		c.Next()
	}
}
