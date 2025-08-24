package repository

//go:generate mockgen -source=async_task_repository.go -destination=./mocks/mock_async_task_repository.go -package=mocks

import (
	"context"

	"gorm.io/gorm"

	"github.com/Haevnen/audit-logging-api/internal/entity/async_task"
)

type AsyncTaskRepository interface {
	Create(ctx context.Context, db *gorm.DB, task *async_task.AsyncTask) (*async_task.AsyncTask, error)
	UpdateStatus(ctx context.Context, db *gorm.DB, taskID string, status async_task.AsyncTaskStatus, errorMsg *string) error
	GetByID(ctx context.Context, taskID string) (*async_task.AsyncTask, error)
}

type asyncTaskRepository struct {
	db *gorm.DB
}

func NewAsyncTaskRepository(db *gorm.DB) *asyncTaskRepository {
	return &asyncTaskRepository{db: db}
}

func (r *asyncTaskRepository) Create(ctx context.Context, db *gorm.DB, task *async_task.AsyncTask) (*async_task.AsyncTask, error) {
	if db == nil {
		db = r.db
	}
	return task, db.WithContext(ctx).Create(task).Error
}

func (r *asyncTaskRepository) UpdateStatus(ctx context.Context, db *gorm.DB, taskID string, status async_task.AsyncTaskStatus, errorMsg *string) error {
	if db == nil {
		db = r.db
	}
	return db.WithContext(ctx).Where("task_id = ?", taskID).Updates(async_task.AsyncTask{Status: status, ErrorMsg: errorMsg}).Error
}

func (r *asyncTaskRepository) GetByID(ctx context.Context, taskID string) (*async_task.AsyncTask, error) {
	var task async_task.AsyncTask
	return &task, r.db.WithContext(ctx).Where("task_id = ?", taskID).First(&task).Error
}
