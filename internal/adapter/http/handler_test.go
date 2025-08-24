package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	h "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
)

func TestHandler_GetPing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h.Handler{}.GetPing(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp api_service.Pong
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "pong", resp.Ping)
}

func TestBindRequestBody_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := []byte(`{"foo":"bar"}`)
	req, _ := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var parsed struct {
		Foo string `json:"foo"`
	}
	err := h.BindRequestBody(c, &parsed)

	assert.NoError(t, err)
	assert.Equal(t, "bar", parsed.Foo)
}

func TestBindRequestBody_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := []byte(`{bad json}`)
	req, _ := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var parsed map[string]any
	err := h.BindRequestBody(c, &parsed)

	assert.Error(t, err)
}

func TestSendError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h.SendError(c, "something went wrong", apperror.ErrInvalidRequestInput)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp api_service.Error
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "something went wrong", resp.Title)
	assert.Equal(t, "validation_failed", string(resp.Type)) // ✅ cast to string
	assert.NotEmpty(t, resp.Code)
	assert.NotEmpty(t, resp.Detail)
}
