package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	h "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	entitytenant "github.com/Haevnen/audit-logging-api/internal/entity/tenant"

	ucMocks "github.com/Haevnen/audit-logging-api/internal/usecase/tenant/mocks"
)

func TestTenantHandler_ListTenants_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockListTenantsUseCaseInterface(ctrl)
	handler := h.TenantHandler{ListUC: mockUC}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// ✅ must set a request
	req, _ := http.NewRequest(http.MethodGet, "/tenants", nil)
	c.Request = req

	now := time.Now()
	expected := []entitytenant.Tenant{
		{ID: "t1", Name: "Tenant1", CreatedAt: now, UpdatedAt: now},
	}

	mockUC.EXPECT().Execute(gomock.Any()).Return(expected, nil)

	handler.ListTenants(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Tenant1")
}

func TestTenantHandler_ListTenants_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockListTenantsUseCaseInterface(ctrl)
	handler := h.TenantHandler{ListUC: mockUC}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// ✅ must set a request
	req, _ := http.NewRequest(http.MethodGet, "/tenants", nil)
	c.Request = req

	mockUC.EXPECT().Execute(gomock.Any()).Return(nil, errors.New("db error"))

	handler.ListTenants(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTenantHandler_CreateTenant_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockCreateTenantUseCaseInterface(ctrl)
	handler := h.TenantHandler{CreateUC: mockUC}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := api_service.CreateTenantRequestBody{Name: "TenantX"}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/tenants", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	now := time.Now()
	expected := &entitytenant.Tenant{ID: "t123", Name: "TenantX", CreatedAt: now, UpdatedAt: now}
	mockUC.EXPECT().Execute(gomock.Any(), "TenantX").Return(expected, nil)

	handler.CreateTenant(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "TenantX")
}

func TestTenantHandler_CreateTenant_InvalidBody(t *testing.T) {
	handler := h.TenantHandler{}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// bad json
	req, _ := http.NewRequest(http.MethodPost, "/tenants", bytes.NewBufferString("{bad json}"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateTenant(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// ✅ updated assertion
	assert.Contains(t, w.Body.String(), "ERR_400")
}
func TestTenantHandler_CreateTenant_MissingName(t *testing.T) {
	handler := h.TenantHandler{}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := api_service.CreateTenantRequestBody{Name: ""}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/tenants", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateTenant(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "name is required")
}

func TestTenantHandler_CreateTenant_FailUC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockCreateTenantUseCaseInterface(ctrl)
	handler := h.TenantHandler{CreateUC: mockUC}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := api_service.CreateTenantRequestBody{Name: "BadTenant"}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/tenants", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	mockUC.EXPECT().Execute(gomock.Any(), "BadTenant").Return(nil, errors.New("db error"))

	handler.CreateTenant(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
