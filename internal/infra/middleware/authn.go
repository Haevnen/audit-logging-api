package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	handler "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/auth"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "Bearer "
)

func RequireAuth(jwtManager *auth.Manager) api_service.MiddlewareFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authHeader) == 0 {
			c.Abort()
			handler.SendError(c, "authentication header not provided", apperror.ErrNotProvidedAuthenticationHeader)
			return
		}
		if !strings.HasPrefix(authHeader, AuthorizationTypeBearer) {
			c.Abort()
			handler.SendError(c, "not supported authorization type", apperror.ErrUnsupportedAuthorizationType)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, AuthorizationTypeBearer)
		claims, err := jwtManager.ParseToken(tokenStr)
		if err != nil {
			c.Abort()
			handler.SendError(c, "invalid token", apperror.ErrInvalidToken)
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
