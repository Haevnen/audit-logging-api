package middleware

import (
	handler "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...auth.Role) api_service.MiddlewareFunc {
	return func(c *gin.Context) {
		role := c.MustGet("role").(auth.Role)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		c.Abort()
		handler.SendError(c, "forbidden", apperror.ErrForbidden)
	}
}
