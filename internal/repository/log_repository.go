package repository

import (
	"context"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"gorm.io/gorm"
)

const (
	CreateBatchSize = 300
)

type LogRepository interface {
	Create(ctx context.Context, log *log.Log) error
	CreateBulk(ctx context.Context, db *gorm.DB, logs []log.Log) error
	GetByID(ctx context.Context, id string, tenant_id string) (*log.Log, error)
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

func (r *logRepository) GetByID(ctx context.Context, id string, tenant_id string) (*log.Log, error) {
	var log log.Log
	q := r.db.WithContext(ctx).Where("id = ?", id)

	if len(tenant_id) > 0 {
		// user or auditor
		q = q.Where("tenant_id = ?", tenant_id)
	}
	err := q.First(&log).Error
	return &log, err
}
