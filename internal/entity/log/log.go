package log

import (
	"time"

	"gorm.io/datatypes"
)

// Severity levels
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityError    Severity = "ERROR"
	SeverityCritical Severity = "CRITICAL"
)

type ActionType string

const (
	ActionCreate ActionType = "CREATE"
	ActionUpdate ActionType = "UPDATE"
	ActionDelete ActionType = "DELETE"
	ActionView   ActionType = "VIEW"
)

// Log entry model
type Log struct {
	// required
	ID             string
	TenantID       string
	UserID         string
	Action         ActionType
	Severity       Severity
	EventTimestamp time.Time

	// optional
	Message    *string
	SessionID  *string
	Resource   *string
	ResourceID *string
	IPAddress  *string
	UserAgent  *string
	Before     *datatypes.JSON
	After      *datatypes.JSON
	Metadata   *datatypes.JSON
}
