package middleware

import (
	"net/http"
	"strings"

	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// ContextKeyUserID is the gin context key for the authenticated user ID.
	ContextKeyUserID = "user_id"
	// ContextKeyEmail is the gin context key for the authenticated user email.
	ContextKeyEmail = "email"
	// ContextKeyRole is the gin context key for the authenticated user role.
	ContextKeyRole = "role"
)

// AuthMiddleware creates a JWT authentication middleware.
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header must be Bearer {token}",
			})
			return
		}

		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// RequireRole creates middleware that restricts access to specific roles.
func RequireRole(roles ...auth.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		userRole, ok := roleVal.(auth.UserRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid role in context",
			})
			return
		}

		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions",
		})
	}
}

// GetUserID extracts the user ID from the gin context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.UUID{}, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetUserRole extracts the user role from the gin context.
func GetUserRole(c *gin.Context) (auth.UserRole, bool) {
	val, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}
	role, ok := val.(auth.UserRole)
	return role, ok
}
