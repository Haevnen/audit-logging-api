package middleware

import (
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	handler "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/constant"
)

var tenantLimiter sync.Map

// Use token bucket algorithm
func getLimiter(tenantID string, reqPerSec, burst int) *rate.Limiter {
	val, ok := tenantLimiter.Load(tenantID)
	if ok {
		return val.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(rate.Limit(reqPerSec), burst)
	tenantLimiter.Store(tenantID, limiter)
	return limiter
}

func RequireRateLimit(reqPerSec, burst int) api_service.MiddlewareFunc {
	return func(c *gin.Context) {
		key := c.Request.Method + ":" + strings.TrimPrefix(c.FullPath(), constant.BaseURL)
		if key == exceptionAPI {
			c.Next()
			return
		}

		tenantID, role := c.GetString(constant.TenantID), c.MustGet(constant.Role).(auth.Role)
		if role == auth.RoleAdmin {
			// Admin has no rate limit
			c.Next()
			return
		}

		if tenantID == "" {
			c.Abort()
			handler.SendError(c, "tenant id not provided", apperror.ErrInvalidToken)
			return
		}

		limiter := getLimiter(tenantID, reqPerSec, burst)
		if !limiter.Allow() {
			c.Abort()
			handler.SendError(c, "rate limit exceed", apperror.ErrTooManyRequests)
			return
		}
		c.Next()
	}
}
