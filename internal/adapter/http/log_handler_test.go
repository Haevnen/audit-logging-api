package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	h "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/constant"
	entitylog "github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/repository"

	ucMocks "github.com/Haevnen/audit-logging-api/internal/usecase/log/mocks"
)

func setupContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set(constant.UserID, "user-1")
	c.Set(constant.TenantID, "tenant-1")
	c.Set(constant.Role, auth.RoleUser) // ✅ use correct type
	return c, w
}

func TestLogHandler_CreateLog_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockCreateLogUseCaseInterface(ctrl)
	handler := h.LogHandler{CreateUC: mockUC}

	body := api_service.CreateLogRequestBody{
		TenantId: "tenant-1", UserId: "user-1", Action: "CREATE", Severity: "INFO", // ✅ fixed
	}
	data, _ := json.Marshal(body)
	c, w := setupContext(http.MethodPost, "/logs", data)

	expected := &entitylog.Log{ID: "log-123", EventTimestamp: time.Now().UTC()}
	mockUC.EXPECT().Execute(gomock.Any(), "tenant-1", "user-1", gomock.Any()).
		Return(expected, nil)

	handler.CreateLog(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "log-123")
}

func TestLogHandler_CreateBulkLogs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockCreateLogUseCaseInterface(ctrl)
	handler := h.LogHandler{CreateUC: mockUC}

	bodies := []api_service.CreateLogRequestBody{{
		TenantId: "tenant-1", UserId: "user-1", Action: "CREATE", Severity: "INFO", // ✅ fixed
	}}
	data, _ := json.Marshal(bodies)
	c, w := setupContext(http.MethodPost, "/logs/bulk", data)

	expected := []entitylog.Log{{ID: "bulk-1", EventTimestamp: time.Now().UTC()}}
	mockUC.EXPECT().ExecuteBulk(gomock.Any(), "tenant-1", "user-1", gomock.Any()).
		Return(expected, nil)

	handler.CreateBulkLogs(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "bulk-1")
}

func TestLogHandler_GetLog_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockGetLogUseCaseInterface(ctrl)
	handler := h.LogHandler{GetUC: mockUC}

	c, w := setupContext(http.MethodGet, "/logs/id-1", nil)
	expected := &entitylog.Log{ID: "id-1"}
	mockUC.EXPECT().Execute(gomock.Any(), "id-1", "tenant-1").Return(expected, nil)

	handler.GetLog(c, "id-1")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "id-1")
}

func TestLogHandler_GetLog_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockGetLogUseCaseInterface(ctrl)
	handler := h.LogHandler{GetUC: mockUC}

	c, w := setupContext(http.MethodGet, "/logs/id-1", nil)
	mockUC.EXPECT().Execute(gomock.Any(), "id-1", "tenant-1").Return(nil, gorm.ErrRecordNotFound)

	handler.GetLog(c, "id-1")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLogHandler_CleanupLogs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockDeleteLogUseCaseInterface(ctrl)
	handler := h.LogHandler{DeleteUC: mockUC}

	c, w := setupContext(http.MethodDelete, "/logs/cleanup", nil)
	params := api_service.CleanupLogsParams{BeforeDate: time.Now().Add(-24 * time.Hour)}

	mockUC.EXPECT().Execute(gomock.Any(), "tenant-1", "user-1", params.BeforeDate).Return(nil)

	handler.CleanupLogs(c, params)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "cleanup successful")
}

func TestLogHandler_GetLogsStat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockGetStatsUseCaseInterface(ctrl)
	handler := h.LogHandler{StatsUC: mockUC}

	c, w := setupContext(http.MethodGet, "/logs/stat", nil)
	params := api_service.GetLogsStatParams{StartDate: time.Now().Add(-24 * time.Hour)}

	mockUC.EXPECT().
		Execute(gomock.Any(), "tenant-1", params.StartDate, gomock.Any()).
		Return([]entitylog.LogStats{}, nil)

	handler.GetLogsStat(c, params)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogHandler_GetLogsStat_InvalidRange(t *testing.T) {
	handler := h.LogHandler{}

	c, w := setupContext(http.MethodGet, "/logs/stat", nil)
	start := time.Now()
	end := start.Add(-time.Hour)
	params := api_service.GetLogsStatParams{StartDate: start, EndDate: &end}

	handler.GetLogsStat(c, params)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogHandler_SearchLogs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockSearchLogsUseCaseInterface(ctrl)
	handler := h.LogHandler{SearchLogUC: mockUC}

	c, w := setupContext(http.MethodGet, "/logs/search", nil)
	params := api_service.SearchLogsParams{}
	expected := &repository.SearchResult{Total: 1, Logs: []entitylog.Log{{ID: "log-1"}}}

	mockUC.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(expected, nil)

	handler.SearchLogs(c, params)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "log-1")
}

func TestLogHandler_SearchLogs_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockSearchLogsUseCaseInterface(ctrl)
	handler := h.LogHandler{SearchLogUC: mockUC}

	c, w := setupContext(http.MethodGet, "/logs/search", nil)
	params := api_service.SearchLogsParams{}

	mockUC.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

	handler.SearchLogs(c, params)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLogHandler_ExportLogs_JSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockSearchLogsUseCaseInterface(ctrl)
	handler := h.LogHandler{SearchLogUC: mockUC}

	c, w := setupContext(http.MethodGet, "/logs/export", nil)
	params := api_service.ExportLogsParams{Format: "json"}

	mockUC.EXPECT().
		Stream(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ repository.LogSearchFilters, fn func(entitylog.Log) error) error {
			return fn(entitylog.Log{ID: "log-1"})
		})

	handler.ExportLogs(c, params)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "log-1")
}

func TestLogHandler_ExportLogs_CSV(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := ucMocks.NewMockSearchLogsUseCaseInterface(ctrl)
	handler := h.LogHandler{SearchLogUC: mockUC}

	c, w := setupContext(http.MethodGet, "/logs/export", nil)
	params := api_service.ExportLogsParams{Format: "csv"}

	mockUC.EXPECT().
		Stream(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ repository.LogSearchFilters, fn func(entitylog.Log) error) error {
			return fn(entitylog.Log{
				ID: "log-1", TenantID: "t1", UserID: "u1",
				Action: "CREATE", Severity: "INFO", // ✅ fixed
				EventTimestamp: time.Now(), Message: "msg",
			})
		})

	handler.ExportLogs(c, params)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "msg")
}

func TestLogHandler_ExportLogs_BadFormat(t *testing.T) {
	handler := h.LogHandler{}

	c, w := setupContext(http.MethodGet, "/logs/export", nil)
	params := api_service.ExportLogsParams{Format: "bad"}

	handler.ExportLogs(c, params)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
