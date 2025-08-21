package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	handler "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/constant"
)

const (
	exceptionAPI = "POST:/auth/token"
)

var roleMap = map[string][]auth.Role{
	"GET:/logs":            {auth.RoleAdmin, auth.RoleAuditor, auth.RoleUser},
	"POST:/logs":           {auth.RoleAdmin, auth.RoleUser},
	"GET:/logs/:id":        {auth.RoleAdmin, auth.RoleAuditor, auth.RoleUser},
	"GET:/logs/export":     {auth.RoleAdmin, auth.RoleAuditor},
	"GET:/logs/stats":      {auth.RoleAdmin, auth.RoleAuditor, auth.RoleUser},
	"POST:/logs/bulk":      {auth.RoleAdmin, auth.RoleUser},
	"DELETE:/logs/cleanup": {auth.RoleAdmin, auth.RoleUser},
	"WS:/logs/stream":      {auth.RoleAdmin, auth.RoleAuditor, auth.RoleUser},
	"GET:/tenants":         {auth.RoleAdmin},
	"POST:/tenants":        {auth.RoleAdmin},
}

func RequireAuth(jwtManager *auth.Manager) api_service.MiddlewareFunc {
	return func(c *gin.Context) {
		key := c.Request.Method + ":" + strings.TrimPrefix(c.FullPath(), constant.BaseURL)
		if key == exceptionAPI {
			c.Next()
			return
		}

		authHeader := c.GetHeader(constant.AuthorizationHeaderKey)
		if len(authHeader) == 0 {
			c.Abort()
			handler.SendError(c, "authentication header not provided", apperror.ErrNotProvidedAuthenticationHeader)
			return
		}
		if !strings.HasPrefix(authHeader, constant.AuthorizationTypeBearer) {
			c.Abort()
			handler.SendError(c, "not supported authorization type", apperror.ErrUnsupportedAuthorizationType)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, constant.AuthorizationTypeBearer)
		claims, err := jwtManager.ParseToken(tokenStr)
		if err != nil {
			c.Abort()
			handler.SendError(c, "invalid token", apperror.ErrInvalidToken)
			return
		}

		c.Set(constant.UserID, claims.UserID)
		c.Set(constant.TenantID, claims.TenantID)
		c.Set(constant.Role, claims.Role)

		c.Next()
	}
}

func RequireRole() api_service.MiddlewareFunc {
	return func(c *gin.Context) {
		key := c.Request.Method + ":" + strings.TrimPrefix(c.FullPath(), constant.BaseURL)
		if key == exceptionAPI {
			c.Next()
			return
		}

		allowedRoles, ok := roleMap[key]
		if !ok {
			// no rule defined -> forbid
			c.Abort()
			handler.SendError(c, "forbidden", apperror.ErrForbidden)
			return
		}

		role := c.MustGet(constant.Role).(auth.Role)
		for _, r := range allowedRoles {
			if role == r {
				c.Next()
				return
			}
		}
		c.Abort()
		handler.SendError(c, "forbidden", apperror.ErrForbidden)
	}
}
