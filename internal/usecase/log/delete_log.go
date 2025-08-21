package log

import (
	"context"
	"time"

	"github.com/Haevnen/audit-logging-api/internal/entity/async_task"
	"github.com/Haevnen/audit-logging-api/internal/interactor"
	"github.com/Haevnen/audit-logging-api/internal/repository"
	"github.com/Haevnen/audit-logging-api/internal/service"
	"github.com/google/uuid"
)

type DeleteLogUseCase struct {
	AsyncTaskRepo  repository.AsyncTaskRepository
	QueuePublisher service.SQSPublisher
	TxManager      *interactor.TxManager
}

func NewDeleteLogUseCase(asyncTaskRepo repository.AsyncTaskRepository, queuePublisher service.SQSPublisher, txManager *interactor.TxManager) *DeleteLogUseCase {
	return &DeleteLogUseCase{
		AsyncTaskRepo:  asyncTaskRepo,
		QueuePublisher: queuePublisher,
		TxManager:      txManager,
	}
}

func (uc *DeleteLogUseCase) Execute(ctx context.Context, tenantId, userId string, beforeDate time.Time) error {
	// We need to create an entry in asyncTask table and publish message to
	// archival log queue in one transaction

	return uc.TxManager.TransactionExec(ctx, func(txCtx context.Context) error {
		db := uc.TxManager.GetTx(txCtx)

		task := &async_task.AsyncTask{
			TaskID:   uuid.New().String(),
			TaskType: async_task.TaskArchive,
			Status:   async_task.StatusPending,
			UserID:   userId,
		}

		if len(tenantId) > 0 {
			task.TenantUID = &tenantId
		}

		task, err := uc.AsyncTaskRepo.Create(txCtx, db, task)
		if err != nil {
			return err
		}

		return uc.QueuePublisher.PublishArchiveMessage(txCtx, task.TaskID, beforeDate)
	})
}
