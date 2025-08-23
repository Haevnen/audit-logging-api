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
	Message        string

	// optional
	SessionID   *string
	Resource    *string
	ResourceID  *string
	IPAddress   *string
	UserAgent   *string
	BeforeState *datatypes.JSON
	AfterState  *datatypes.JSON
	Metadata    *datatypes.JSON
}

type LogStats struct {
	Day              time.Time
	Total            int64
	ActionCreate     int64
	ActionUpdate     int64
	ActionDelete     int64
	ActionView       int64
	SeverityInfo     int64
	SeverityError    int64
	SeverityWarning  int64
	SeverityCritical int64
}
