package async_task

import (
	"time"

	"gorm.io/datatypes"
)

type AsyncTaskStatus string
type AsyncTaskType string

const (
	StatusPending   AsyncTaskStatus = "pending"
	StatusRunning   AsyncTaskStatus = "running"
	StatusSucceeded AsyncTaskStatus = "succeeded"
	StatusFailed    AsyncTaskStatus = "failed"

	TaskLogCleanup AsyncTaskType = "log_cleanup"
	TaskArchive    AsyncTaskType = "archive"
	TaskExport     AsyncTaskType = "export"
	TaskReindex    AsyncTaskType = "reindex"
)

type AsyncTask struct {
	TaskID    string // UUID
	Status    AsyncTaskStatus
	TaskType  AsyncTaskType
	Payload   *datatypes.JSON
	CreatedAt time.Time
	UpdatedAt time.Time
	TenantUID *string
	UserID    string
	ErrorMsg  *string
}
