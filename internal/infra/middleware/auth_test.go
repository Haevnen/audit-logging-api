package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/Haevnen/audit-logging-api/internal/auth"
	authMocks "github.com/Haevnen/audit-logging-api/internal/auth/mocks"
	"github.com/Haevnen/audit-logging-api/internal/constant"
	m "github.com/Haevnen/audit-logging-api/internal/infra/middleware"
)

func runRequest(r *gin.Engine, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestRequireAuth_NoHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtMock := authMocks.NewMockManagerInterface(ctrl)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/logs",
		gin.HandlerFunc(m.RequireAuth(jwtMock)), // ✅ cast
		func(c *gin.Context) { c.String(200, "ok") })

	w := runRequest(r, "GET", "/api/v1/logs", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authentication header not provided")
}

func TestRequireAuth_BadAuthType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtMock := authMocks.NewMockManagerInterface(ctrl)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/logs",
		gin.HandlerFunc(m.RequireAuth(jwtMock)), // ✅ cast
		func(c *gin.Context) { c.String(200, "ok") })

	headers := map[string]string{constant.AuthorizationHeaderKey: "Basic sometoken"}
	w := runRequest(r, "GET", "/api/v1/logs", headers)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not supported authorization type")
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtMock := authMocks.NewMockManagerInterface(ctrl)
	jwtMock.EXPECT().ParseToken("badtoken").Return(nil, errors.New("invalid"))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/logs",
		gin.HandlerFunc(m.RequireAuth(jwtMock)), // ✅ cast
		func(c *gin.Context) { c.String(200, "ok") })

	headers := map[string]string{constant.AuthorizationHeaderKey: "Bearer badtoken"}
	w := runRequest(r, "GET", "/api/v1/logs", headers)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}

func TestRequireAuth_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	claims := &auth.Claims{UserID: "u1", TenantID: "t1", Role: auth.RoleUser}
	jwtMock := authMocks.NewMockManagerInterface(ctrl)
	jwtMock.EXPECT().ParseToken("goodtoken").Return(claims, nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/logs",
		gin.HandlerFunc(m.RequireAuth(jwtMock)), // ✅ cast
		func(c *gin.Context) { c.String(200, "ok") })

	headers := map[string]string{constant.AuthorizationHeaderKey: "Bearer goodtoken"}
	w := runRequest(r, "GET", "/api/v1/logs", headers)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func makeRoleRouter(method, path string, role auth.Role) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// inject role
	r.Use(func(c *gin.Context) {
		c.Set(constant.Role, role)
		c.Next()
	})

	r.Handle(method, path,
		gin.HandlerFunc(m.RequireRole()),
		func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		},
	)
	return r
}

func TestRequireRole_ExceptionAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// inject role
	r.Use(func(c *gin.Context) {
		c.Set(constant.Role, auth.RoleUser)
		c.Next()
	})

	// /auth/token → exceptionAPI
	r.POST("/api/v1/auth/token",
		gin.HandlerFunc(m.RequireRole()),
		func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/token", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestRequireRole_NoRuleDefined(t *testing.T) {
	r := makeRoleRouter("GET", "/api/v1/unknown", auth.RoleUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/unknown", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

func TestRequireRole_AllowedRole(t *testing.T) {
	r := makeRoleRouter("GET", "/api/v1/logs", auth.RoleUser) // RoleUser is allowed

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/logs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestRequireRole_NotAllowedRole(t *testing.T) {
	r := makeRoleRouter("GET", "/api/v1/tenants", auth.RoleUser) // Only Admin allowed

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tenants", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}
