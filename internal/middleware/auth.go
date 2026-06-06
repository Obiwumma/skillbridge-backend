// Package middleware intercepts HTTP requests to configure filters, authorization, and rate-limits.
package middleware

import (
	"strings"
	"skillbridge-backend/internal/auth"
	"skillbridge-backend/internal/users"
	"skillbridge-backend/pkg/errors"
	"skillbridge-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware decodes Bearer JWT access tokens and injects claims into the request context.
func AuthMiddleware(authService auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errors.NewUnauthorizedError("authorization header required", nil))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errors.NewUnauthorizedError("invalid authorization scheme", nil))
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := authService.ValidateToken(tokenStr)
		if err != nil {
			response.Error(c, errors.NewUnauthorizedError("invalid or expired authentication token", err))
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

// RequireRole restricts access to endpoints by verifying that the authenticated user matches allowed roles.
func RequireRole(allowedRoles ...users.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("userRole")
		if !exists {
			response.Error(c, errors.NewUnauthorizedError("unauthenticated request session", nil))
			c.Abort()
			return
		}

		roleStr := roleVal.(string)
		isAllowed := false
		for _, role := range allowedRoles {
			if string(role) == roleStr {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			response.Error(c, errors.NewForbiddenError("insufficient privilege level to access this resource", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}
