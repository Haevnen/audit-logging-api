package handler

import (
	"encoding/json"

	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	entity_log "github.com/Haevnen/audit-logging-api/internal/entity/log"
	"gorm.io/datatypes"
)

// Convert from OpenAPI Action → Domain ActionType
func toEntityAction(a api_service.Action) log.ActionType {
	switch a {
	case api_service.CREATE:
		return log.ActionCreate
	case api_service.UPDATE:
		return log.ActionUpdate
	case api_service.DELETE:
		return log.ActionDelete
	case api_service.VIEW:
		return log.ActionView
	default:
		return "" // or panic, but safer to return empty
	}
}

// Same idea for Severity
func toEntitySeverity(s api_service.Severity) log.Severity {
	switch s {
	case api_service.CRITICAL:
		return log.SeverityCritical
	case api_service.ERROR:
		return log.SeverityError
	case api_service.WARNING:
		return log.SeverityWarning
	case api_service.INFO:
		return log.SeverityInfo
	default:
		return ""
	}
}

func MarshallData(data *map[string]interface{}) (*datatypes.JSON, error) {
	if data == nil {
		return nil, nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	d := datatypes.JSON(jsonBytes)
	return &d, nil
}
func JSONToMap(j *datatypes.JSON) (*map[string]interface{}, error) {
	if j == nil || len(*j) == 0 {
		// No value stored in DB
		return nil, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(*j, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func toLogStatsResponse(stats []log.LogStats) []api_service.LogStat {
	var resp []api_service.LogStat
	for _, s := range stats {
		resp = append(resp, api_service.LogStat{
			Day:      s.Day.Format("2006-01-02"),
			Total:    s.Total,
			CREATE:   s.ActionCreate,
			UPDATE:   s.ActionUpdate,
			DELETE:   s.ActionDelete,
			VIEW:     s.ActionView,
			INFO:     s.SeverityInfo,
			ERROR:    s.SeverityError,
			WARNING:  s.SeverityWarning,
			CRITICAL: s.SeverityCritical,
		})
	}
	return resp
}

func toSingleLogResponse(l entity_log.Log) (api_service.GetSingleLogResponse, error) {
	before, err := JSONToMap(l.Before)
	if err != nil {
		return api_service.GetSingleLogResponse{}, err
	}
	after, err := JSONToMap(l.After)
	if err != nil {
		return api_service.GetSingleLogResponse{}, err
	}

	metadata, err := JSONToMap(l.Metadata)
	if err != nil {
		return api_service.GetSingleLogResponse{}, err
	}

	return api_service.GetSingleLogResponse{
		Id:             l.ID,
		UserId:         l.UserID,
		TenantId:       l.TenantID,
		Action:         api_service.Action(l.Action),
		Severity:       api_service.Severity(l.Severity),
		EventTimestamp: l.EventTimestamp.Format(DateTimeFormat),
		Message:        l.Message,
		SessionId:      l.SessionID,
		Resource:       l.Resource,
		ResourceId:     l.ResourceID,
		IpAddress:      l.IPAddress,
		UserAgent:      l.UserAgent,
		Before:         before,
		After:          after,
		Metadata:       metadata,
	}, nil
}
