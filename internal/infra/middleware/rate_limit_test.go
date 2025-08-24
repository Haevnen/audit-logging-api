package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/constant"
	m "github.com/Haevnen/audit-logging-api/internal/infra/middleware"
)

func makeRouter(role auth.Role, tenantID string, reqPerSec, burst int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// inject role + tenant into context before rate limit middleware
	r.Use(func(c *gin.Context) {
		c.Set(constant.Role, role)
		c.Set(constant.TenantID, tenantID)
		c.Next()
	})

	r.GET("/api/v1/logs",
		gin.HandlerFunc(m.RequireRateLimit(reqPerSec, burst)),
		func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		},
	)
	return r
}

func TestRequireRateLimit_AdminBypass(t *testing.T) {
	r := makeRouter(auth.RoleAdmin, "tenant-1", 1, 1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/logs", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestRequireRateLimit_NoTenantID(t *testing.T) {
	r := makeRouter(auth.RoleUser, "", 1, 1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/logs", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "tenant id not provided")
}

func TestRequireRateLimit_Allowed(t *testing.T) {
	r := makeRouter(auth.RoleUser, "tenant-1", 1, 1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/logs", nil)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestRequireRateLimit_Exceed(t *testing.T) {
	r := makeRouter(auth.RoleUser, "tenant-2", 1, 1)

	// first request passes
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/v1/logs", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// second request should hit limiter
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/logs", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Contains(t, w2.Body.String(), "rate limit exceed")
}
