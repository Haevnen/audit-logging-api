package repository

import (
	"context"
	"time"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"gorm.io/gorm"
)

const (
	CreateBatchSize = 300
)

type LogRepository interface {
	Create(ctx context.Context, log *log.Log) error
	CreateBulk(ctx context.Context, db *gorm.DB, logs []log.Log) error
	GetByID(ctx context.Context, id string, tenantId string) (*log.Log, error)
	FindLogsForArchival(ctx context.Context, tenantId *string, beforeDate time.Time) ([]log.Log, error)
	CleanupLogsBefore(ctx context.Context, db *gorm.DB, tenantId *string, beforeDate time.Time) error
	GetStats(ctx context.Context, tenantId string, startTime, endTime time.Time) ([]log.LogStats, error)
}

type logRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) *logRepository {
	return &logRepository{db: db}
}

func (r *logRepository) Create(ctx context.Context, log *log.Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *logRepository) CreateBulk(ctx context.Context, db *gorm.DB, logs []log.Log) error {
	if db == nil {
		db = r.db
	}
	return db.WithContext(ctx).CreateInBatches(logs, CreateBatchSize).Error
}

func (r *logRepository) GetByID(ctx context.Context, id string, tenantId string) (*log.Log, error) {
	var log log.Log
	q := r.db.WithContext(ctx).Where("id = ?", id)

	if len(tenantId) > 0 {
		// user or auditor
		q = q.Where("tenant_id = ?", tenantId)
	}
	err := q.First(&log).Error
	return &log, err
}

func (r *logRepository) FindLogsForArchival(ctx context.Context, tenantID *string, beforeDate time.Time) ([]log.Log, error) {
	allLogs := make([]log.Log, 0)
	batchSize := 1000 // tune based on your workload

	q := r.db.WithContext(ctx).Where("event_timestamp < ?", beforeDate)
	if tenantID != nil && len(*tenantID) > 0 {
		q = q.Where("tenant_id = ?", *tenantID)
	}

	err := q.FindInBatches(&[]log.Log{}, batchSize, func(tx *gorm.DB, batch int) error {
		var batchLogs []log.Log
		if err := tx.Find(&batchLogs).Error; err != nil {
			return err
		}
		allLogs = append(allLogs, batchLogs...)
		return nil
	}).Error

	if err != nil {
		return nil, err
	}
	return allLogs, nil
}

func (r *logRepository) CleanupLogsBefore(ctx context.Context, db *gorm.DB, tenantId *string, beforeDate time.Time) error {
	if db == nil {
		db = r.db
	}

	q := db.WithContext(ctx).Where("event_timestamp < ?", beforeDate)
	if tenantId != nil && len(*tenantId) > 0 {
		q = q.Where("tenant_id = ?", *tenantId)
	}
	return q.Delete(&log.Log{}).Error
}

func (r *logRepository) GetStats(ctx context.Context, tenantId string, startTime, endTime time.Time) ([]log.LogStats, error) {
	var stats []log.LogStats

	if len(tenantId) > 0 {
		query := `
        SELECT
			day,
			SUM(log_count) AS total_count,
			SUM(CASE WHEN action = 'CREATE' THEN log_count ELSE 0 END) AS create_count,
			SUM(CASE WHEN action = 'UPDATE' THEN log_count ELSE 0 END) AS update_count,
			SUM(CASE WHEN action = 'DELETE' THEN log_count ELSE 0 END) AS delete_count,
			SUM(CASE WHEN action = 'VIEW'   THEN log_count ELSE 0 END) AS view_count,
			SUM(CASE WHEN severity = 'INFO'     THEN log_count ELSE 0 END) AS info_count,
			SUM(CASE WHEN severity = 'ERROR'    THEN log_count ELSE 0 END) AS error_count,
			SUM(CASE WHEN severity = 'WARNING'  THEN log_count ELSE 0 END) AS warning_count,
			SUM(CASE WHEN severity = 'CRITICAL' THEN log_count ELSE 0 END) AS critical_count
		FROM log_stats_daily
		WHERE tenant_id = ? AND day BETWEEN ? AND ?
		GROUP BY day
		ORDER BY day DESC;`
		err := r.db.WithContext(ctx).Raw(query, tenantId, startTime, endTime).Scan(&stats).Error
		return stats, err
	}

	query := `
        SELECT
			day,
			SUM(log_count) AS total,
			SUM(CASE WHEN action = 'CREATE' THEN log_count ELSE 0 END) AS action_create,
			SUM(CASE WHEN action = 'UPDATE' THEN log_count ELSE 0 END) AS action_update,
			SUM(CASE WHEN action = 'DELETE' THEN log_count ELSE 0 END) AS action_delete,
			SUM(CASE WHEN action = 'VIEW'   THEN log_count ELSE 0 END) AS action_view,
			SUM(CASE WHEN severity = 'INFO'     THEN log_count ELSE 0 END) AS severity_info,
			SUM(CASE WHEN severity = 'ERROR'    THEN log_count ELSE 0 END) AS severity_error,
			SUM(CASE WHEN severity = 'WARNING'  THEN log_count ELSE 0 END) AS severity_warning,
			SUM(CASE WHEN severity = 'CRITICAL' THEN log_count ELSE 0 END) AS severity_critical
		FROM log_stats_daily
		WHERE day BETWEEN ? AND ?
		GROUP BY day
		ORDER BY day DESC;`
	err := r.db.WithContext(ctx).Raw(query, startTime, endTime).Scan(&stats).Error
	return stats, err

}
