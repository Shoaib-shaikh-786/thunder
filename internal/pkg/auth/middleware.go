package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var publicPaths = []string{
	"/health",
	"/auth/login",
	"/auth/role",
	"/auth/dealer/join",
}

var publicPathPrefixes = []string{
	"/api/mock",
}

func isPublicPath(path string) bool {
	for _, p := range publicPaths {
		if path == p {
			return true
		}
	}
	for _, prefix := range publicPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func AuthMiddleware(authService BaseAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicPath(c.Request.URL.Path) || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		claim, err := authService.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_claim", claim)
		c.Next()
	}
}
