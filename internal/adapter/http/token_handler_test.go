package handler_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	h "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	"github.com/Haevnen/audit-logging-api/internal/auth"

	authMocks "github.com/Haevnen/audit-logging-api/internal/auth/mocks"
)

func TestTokenHandler_GenerateToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := authMocks.NewMockManagerInterface(ctrl)
	handler := h.TokenHandler{JwtManager: mockManager}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// ✅ lowercase role "user"
	body := []byte(`{"user_id":"u1","tenant_id":"t1","role":"user"}`)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockManager.EXPECT().
		GenerateToken("u1", "t1", auth.RoleUser, 1*time.Hour).
		Return("fake-token", nil)

	handler.GenerateToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "fake-token")
}

func TestTokenHandler_GenerateToken_InvalidBody(t *testing.T) {
	handler := h.TokenHandler{}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// malformed json
	body := []byte(`{"user_id":123}`)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GenerateToken(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestTokenHandler_GenerateToken_InvalidRole(t *testing.T) {
	handler := h.TokenHandler{}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// invalid role
	body := []byte(`{"user_id":"u1","tenant_id":"t1","role":"INVALID"}`)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GenerateToken(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid role")
}

func TestTokenHandler_GenerateToken_JWTError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := authMocks.NewMockManagerInterface(ctrl)
	handler := h.TokenHandler{JwtManager: mockManager}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// ✅ lowercase role "user"
	body := []byte(`{"user_id":"u1","tenant_id":"t1","role":"user"}`)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockManager.EXPECT().
		GenerateToken("u1", "t1", auth.RoleUser, 1*time.Hour).
		Return("", errors.New("jwt error"))

	handler.GenerateToken(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to generate token")
}
