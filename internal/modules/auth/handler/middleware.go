package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"recipes-desk/internal/modules/auth/service"
)

type tokenService interface {
	ValidateToken(tokenString string) (*service.Claims, error)
}

// AuthMiddleware creates a middleware that validates JWT tokens from cookies.
func AuthMiddleware(service tokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token from cookie first
		tokenString, err := c.Cookie(accessTokenCookieName)
		if err != nil {
			// Fallback to Authorization header (Bearer token)
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
				c.Abort()
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				c.JSON(http.StatusUnauthorized,
					gin.H{"error": "invalid authorization header format"})
				c.Abort()
				return
			}
			tokenString = parts[1]
		}

		// Validate token
		claims, err := service.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)

		c.Next()
	}
}
