package middleware

import (
	"net/http"
	"strings"

	"github.com/Prathyusha2909/quantumfield/internal/auth"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "user_id"
	ContextRole   = "role"
)

func Authenticate(authService *auth.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		header := context.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := authService.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		context.Set(ContextUserID, claims.UserID)
		context.Set(ContextRole, claims.Role)
		context.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(context *gin.Context) {
		role := context.GetString(ContextRole)
		if !allowed[role] {
			context.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		context.Next()
	}
}
