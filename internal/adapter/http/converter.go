package handler

import (
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/entity/log"
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
