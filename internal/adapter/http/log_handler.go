package handler

import (
	"errors"
	"net/http"

	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/constant"
	entity_log "github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/Haevnen/audit-logging-api/internal/usecase/log"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type logHandler struct {
	CreateUC *log.CreateLogUseCase
	GetUC    *log.GetLogUseCase
	DeleteUC *log.DeleteLogUseCase
}

func newLogHandler(r *registry.Registry) logHandler {
	return logHandler{
		CreateUC: r.CreateLogUseCase(),
		GetUC:    r.GetLogUseCase(),
		DeleteUC: r.DeleteLogUseCase(),
	}
}

// (POST /logs)
func (h logHandler) CreateLog(g *gin.Context) {
	var body api_service.CreateLogRequestBody
	if err := bindRequestBody(g, &body); err != nil {
		SendError(g, err.Error(), apperror.ErrInvalidRequestInput)
		return
	}

	e, title, err := validateAndGenerateLogEntity(g, body)
	if err != nil {
		SendError(g, title, err)
		return
	}

	logCreated, err := h.CreateUC.Execute(g.Request.Context(), e)
	if err != nil {
		SendError(g, err.Error(), apperror.ErrInternalServer)
		return
	}

	resp := api_service.CreateLogResponse{
		Id:             logCreated.ID,
		EventTimestamp: logCreated.EventTimestamp.Format(DateTimeFormat),
	}
	g.JSON(http.StatusCreated, resp)
}

// (POST /logs/bulk)
func (h logHandler) CreateBulkLogs(c *gin.Context) {
	var body []api_service.CreateLogRequestBody
	if err := bindRequestBody(c, &body); err != nil {
		SendError(c, err.Error(), apperror.ErrInvalidRequestInput)
		return
	}

	logs := make([]entity_log.Log, 0, len(body))
	for _, b := range body {
		e, title, err := validateAndGenerateLogEntity(c, b)
		if err != nil {
			SendError(c, title, err)
			return
		}

		logs = append(logs, e)
	}

	logsCreated, err := h.CreateUC.ExecuteBulk(c.Request.Context(), logs)
	if err != nil {
		SendError(c, err.Error(), apperror.ErrInternalServer)
		return
	}

	resp := make([]api_service.CreateLogResponse, 0, len(logsCreated))
	for _, l := range logsCreated {
		resp = append(resp, api_service.CreateLogResponse{
			Id:             l.ID,
			EventTimestamp: l.EventTimestamp.Format(DateTimeFormat),
		})
	}
	c.JSON(http.StatusCreated, resp)
}

// (GET /logs/{id})
func (h logHandler) GetLog(c *gin.Context, id string) {
	if len(id) == 0 {
		SendError(c, "id is required", apperror.ErrInvalidRequestInput)
		return
	}

	claimTenantId := getClaimTenant(c)
	log, err := h.GetUC.Execute(c.Request.Context(), id, claimTenantId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			SendError(c, err.Error(), apperror.ErrRecordNotFound)
			return
		}
		SendError(c, err.Error(), apperror.ErrInternalServer)
		return
	}

	before, err := JSONToMap(log.Before)
	if err != nil {
		SendError(c, err.Error(), apperror.ErrInternalServer)
		return
	}
	after, err := JSONToMap(log.After)
	if err != nil {
		SendError(c, err.Error(), apperror.ErrInternalServer)
		return
	}

	metadata, err := JSONToMap(log.Metadata)
	if err != nil {
		SendError(c, err.Error(), apperror.ErrInternalServer)
		return
	}

	resp := api_service.GetSingleLogResponse{
		Id:             log.ID,
		UserId:         log.UserID,
		TenantId:       log.TenantID,
		Action:         api_service.Action(log.Action),
		Severity:       api_service.Severity(log.Severity),
		EventTimestamp: log.EventTimestamp.Format(DateTimeFormat),
		Message:        log.Message,
		SessionId:      log.SessionID,
		Resource:       log.Resource,
		ResourceId:     log.ResourceID,
		IpAddress:      log.IPAddress,
		UserAgent:      log.UserAgent,
		Before:         before,
		After:          after,
		Metadata:       metadata,
	}

	c.JSON(http.StatusOK, resp)
}

// (DELETE /logs/cleanup)
func (h logHandler) CleanupLogs(c *gin.Context, params api_service.CleanupLogsParams) {
	tenantId := getClaimTenant(c)
	userId := c.GetString(constant.UserID)

	if err := h.DeleteUC.Execute(c.Request.Context(), tenantId, userId, params.BeforeDate); err != nil {
		SendError(c, err.Error(), apperror.ErrInternalServer)
		return
	}

	c.JSON(http.StatusOK, "cleanup successful")
}

func validateAndGenerateLogEntity(g *gin.Context, body api_service.CreateLogRequestBody) (entity_log.Log, string, error) {
	claimTenantId := getClaimTenant(g)

	if err := validateMismatchTenant(claimTenantId, body.TenantId); err != nil {
		return entity_log.Log{}, "tenant id mismatch", err
	}

	// validate required fields
	if len(body.UserId) == 0 {
		return entity_log.Log{}, "user id is required", apperror.ErrInvalidRequestInput
	}

	actionType := toEntityAction(body.Action)
	if actionType == "" {
		return entity_log.Log{}, "invalid action type", apperror.ErrInvalidRequestInput
	}

	severity := toEntitySeverity(body.Severity)
	if severity == "" {
		return entity_log.Log{}, "invalid severity", apperror.ErrInvalidRequestInput
	}

	beforeJSON, err := MarshallData(body.Before)
	if err != nil {
		return entity_log.Log{}, err.Error(), apperror.ErrInvalidRequestInput
	}

	afterJSON, err := MarshallData(body.After)
	if err != nil {
		return entity_log.Log{}, err.Error(), apperror.ErrInvalidRequestInput
	}

	metaDataJSON, err := MarshallData(body.Metadata)
	if err != nil {
		return entity_log.Log{}, err.Error(), apperror.ErrInvalidRequestInput
	}

	return entity_log.Log{
		TenantID:       body.TenantId,
		UserID:         body.UserId,
		SessionID:      body.SessionId,
		Message:        body.Message,
		Action:         actionType,
		Resource:       body.Resource,
		ResourceID:     body.ResourceId,
		Severity:       severity,
		IPAddress:      body.IpAddress,
		UserAgent:      body.UserAgent,
		Before:         beforeJSON,
		After:          afterJSON,
		Metadata:       metaDataJSON,
		EventTimestamp: body.EventTimestamp,
	}, "", nil
}

func getClaimTenant(g *gin.Context) string {
	claimTenantId := g.GetString(constant.TenantID)
	role := g.MustGet(constant.Role).(auth.Role)

	if role == auth.RoleAdmin {
		return ""
	}
	return claimTenantId
}

func validateMismatchTenant(claimTenantId, bodyTenantId string) error {
	if len(claimTenantId) == 0 || claimTenantId == bodyTenantId {
		// admin
		return nil
	}
	return apperror.ErrForbidden
}
