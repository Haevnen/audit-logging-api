package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/constant"
	entity_log "github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/Haevnen/audit-logging-api/internal/repository"
	"github.com/Haevnen/audit-logging-api/internal/usecase/log"
	"github.com/Haevnen/audit-logging-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LogHandler struct {
	CreateUC    log.CreateLogUseCaseInterface
	GetUC       log.GetLogUseCaseInterface
	DeleteUC    log.DeleteLogUseCaseInterface
	StatsUC     log.GetStatsUseCaseInterface
	SearchLogUC log.SearchLogsUseCaseInterface
}

func newLogHandler(r *registry.Registry) LogHandler {
	return LogHandler{
		CreateUC:    r.CreateLogUseCase(),
		GetUC:       r.GetLogUseCase(),
		DeleteUC:    r.DeleteLogUseCase(),
		StatsUC:     r.GetStatsUseCase(),
		SearchLogUC: r.SearchLogsUseCase(),
	}
}

// (POST /logs)
func (h LogHandler) CreateLog(g *gin.Context) {
	userId := g.GetString(constant.UserID)
	tenantId := getClaimTenant(g)

	var body api_service.CreateLogRequestBody
	if err := BindRequestBody(g, &body); err != nil {
		SendError(g, err.Error(), apperror.ErrInvalidRequestInput)
		return
	}

	e, title, err := validateAndGenerateLogEntity(g, body)
	if err != nil {
		SendError(g, title, err)
		return
	}

	logCreated, err := h.CreateUC.Execute(g.Request.Context(), tenantId, userId, e)
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
func (h LogHandler) CreateBulkLogs(c *gin.Context) {
	tenantId := getClaimTenant(c)
	userId := c.GetString(constant.UserID)

	var body []api_service.CreateLogRequestBody
	if err := BindRequestBody(c, &body); err != nil {
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

	logsCreated, err := h.CreateUC.ExecuteBulk(c.Request.Context(), tenantId, userId, logs)
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
func (h LogHandler) GetLog(c *gin.Context, id string) {
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

	resp, err := ToSingleLogResponse(*log)
	if err != nil {
		SendError(c, err.Error(), apperror.ErrInternalServer)
	}
	c.JSON(http.StatusOK, resp)
}

// (DELETE /logs/cleanup)
func (h LogHandler) CleanupLogs(c *gin.Context, params api_service.CleanupLogsParams) {
	tenantId := getClaimTenant(c)
	userId := c.GetString(constant.UserID)

	if err := h.DeleteUC.Execute(c.Request.Context(), tenantId, userId, params.BeforeDate); err != nil {
		SendError(c, err.Error(), apperror.ErrInternalServer)
		return
	}

	c.JSON(http.StatusOK, "cleanup successful")
}

// (GET /logs/stat)
func (h LogHandler) GetLogsStat(c *gin.Context, params api_service.GetLogsStatParams) {
	tenantId := getClaimTenant(c)

	endDate := time.Now().UTC()
	if params.EndDate != nil {
		endDate = *params.EndDate
	}

	if endDate.Before(params.StartDate) {
		SendError(c, "end date must be after start date", apperror.ErrInvalidRequestInput)
		return
	}

	stats, err := h.StatsUC.Execute(c.Request.Context(), tenantId, params.StartDate, endDate)
	if err != nil {
		SendError(c, err.Error(), apperror.ErrInternalServer)
		return
	}

	c.JSON(http.StatusOK, ToLogStatsResponse(stats))
}

// (GET /api/v1/logs/search)
func (h LogHandler) SearchLogs(c *gin.Context, params api_service.SearchLogsParams) {
	pageNumber, pageSize := 1, constant.MaxPageSize
	if params.PageNumber != nil && *params.PageNumber > 0 {
		pageNumber = *params.PageNumber
	}
	if params.PageSize != nil && *params.PageSize > 0 && *params.PageSize <= constant.MaxPageSize {
		pageSize = *params.PageSize
	}

	tenantId := getClaimTenant(c)

	filters := repository.LogSearchFilters{
		TenantID:  utils.Ptr(tenantId),
		UserID:    utils.Ptr(c.Query("user_id")),
		Action:    utils.Ptr(c.Query("action")),
		Resource:  utils.Ptr(c.Query("resource")),
		Severity:  utils.Ptr(c.Query("severity")),
		StartDate: utils.Ptr(c.Query("start_time")),
		EndDate:   utils.Ptr(c.Query("end_time")),
		Query:     utils.Ptr(c.Query("q")),
		Page:      pageNumber,
		PageSize:  pageSize,
	}

	result, err := h.SearchLogUC.Execute(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logConverted := make([]api_service.GetSingleLogResponse, 0, len(result.Logs))
	for _, l := range result.Logs {
		r, err := ToSingleLogResponse(l)
		if err != nil {
			SendError(c, err.Error(), apperror.ErrInternalServer)
			return
		}
		logConverted = append(logConverted, r)
	}

	resp := api_service.InlineResponse200{
		Total:      result.Total,
		Items:      logConverted,
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, resp)
}

// (GET /api/v1/logs/export)
func (h LogHandler) ExportLogs(c *gin.Context, params api_service.ExportLogsParams) {
	ctx := c.Request.Context()

	tenantId := getClaimTenant(c)

	filters := repository.LogSearchFilters{
		TenantID:  utils.Ptr(tenantId),
		UserID:    utils.Ptr(c.Query("user_id")),
		Action:    utils.Ptr(c.Query("action")),
		Resource:  utils.Ptr(c.Query("resource")),
		Severity:  utils.Ptr(c.Query("severity")),
		StartDate: utils.Ptr(c.Query("start_time")),
		EndDate:   utils.Ptr(c.Query("end_time")),
		Query:     utils.Ptr(c.Query("q")),
	}

	format := params.Format

	// prepare HTTP headers
	filename := fmt.Sprintf("logs.%s", format)
	c.Header("Content-Disposition", "attachment; filename="+filename)

	switch format {
	case "json":
		c.Header("Content-Type", "application/json")
		c.Writer.Write([]byte("[")) // start JSON array

		first := true
		err := h.SearchLogUC.Stream(ctx, filters, func(l entity_log.Log) error {
			data, _ := json.Marshal(l)
			if !first {
				c.Writer.Write([]byte(","))
			}
			c.Writer.Write(data)
			first = false
			c.Writer.Flush()
			return nil
		})
		if err != nil {
			SendError(c, err.Error(), apperror.ErrInternalServer)
			return
		}

		c.Writer.Write([]byte("]")) // close JSON array

	case "csv":
		c.Header("Content-Type", "text/csv")
		w := csv.NewWriter(c.Writer)

		// header row
		_ = w.Write([]string{"id", "tenant_id", "user_id", "action", "severity", "event_timestamp", "message"})

		err := h.SearchLogUC.Stream(ctx, filters, func(l entity_log.Log) error {
			return w.Write([]string{
				l.ID,
				l.TenantID,
				l.UserID,
				string(l.Action),
				string(l.Severity),
				l.EventTimestamp.Format(time.RFC3339),
				l.Message,
			})
		})
		w.Flush()

		if err != nil {
			SendError(c, err.Error(), apperror.ErrInternalServer)
			return
		}

	default:
		SendError(c, "bad request", apperror.ErrInvalidRequestInput)
	}
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

	actionType := ToEntityAction(body.Action)
	if actionType == "" {
		return entity_log.Log{}, "invalid action type", apperror.ErrInvalidRequestInput
	}

	severity := ToEntitySeverity(body.Severity)
	if severity == "" {
		return entity_log.Log{}, "invalid severity", apperror.ErrInvalidRequestInput
	}

	beforeJSON, err := MarshallData(body.BeforeState)
	if err != nil {
		return entity_log.Log{}, err.Error(), apperror.ErrInvalidRequestInput
	}

	afterJSON, err := MarshallData(body.AfterState)
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
		BeforeState:    beforeJSON,
		AfterState:     afterJSON,
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
