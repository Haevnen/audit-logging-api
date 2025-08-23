package log

import (
	"context"

	"github.com/Haevnen/audit-logging-api/internal/entity/async_task"
	entitylog "github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/interactor"
	"github.com/Haevnen/audit-logging-api/internal/repository"
	"github.com/Haevnen/audit-logging-api/internal/service"
	"github.com/google/uuid"
)

type CreateLogUseCase struct {
	Repo           repository.LogRepository
	AsyncTaskRepo  repository.AsyncTaskRepository
	TxManager      *interactor.TxManager
	QueuePublisher service.SQSPublisher
	PubSub         service.PubSub
}

func NewCreateLogUseCase(repo repository.LogRepository, txManager *interactor.TxManager, queuePublisher service.SQSPublisher, pubSub service.PubSub, asyncTaskRepo repository.AsyncTaskRepository) *CreateLogUseCase {
	return &CreateLogUseCase{Repo: repo, TxManager: txManager, QueuePublisher: queuePublisher, PubSub: pubSub, AsyncTaskRepo: asyncTaskRepo}
}

func (uc *CreateLogUseCase) Execute(ctx context.Context, tenantId, userId string, log entitylog.Log) (*entitylog.Log, error) {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}

	// Start a transaction to write log to db, publish SQS message to worker to index opensearch
	if err := uc.TxManager.TransactionExec(ctx, func(txCtx context.Context) error {
		db := uc.TxManager.GetTx(txCtx)
		// 1. Write to DB
		if err := uc.Repo.CreateBulk(txCtx, db, []entitylog.Log{log}); err != nil {
			return err
		}

		// 2. Create a corresponding task in asyncTask table
		task := &async_task.AsyncTask{
			TaskID:   uuid.New().String(),
			TaskType: async_task.TaskReindex,
			Status:   async_task.StatusPending,
			UserID:   userId,
		}

		if len(tenantId) > 0 {
			task.TenantUID = &tenantId
		}

		if _, err := uc.AsyncTaskRepo.Create(txCtx, db, task); err != nil {
			return err
		}

		// 3. Publish message to SQS
		return uc.QueuePublisher.PublishIndexMessage(txCtx, task.TaskID, []entitylog.Log{log})
	}); err != nil {
		return nil, err
	}

	// Broadcast log to redis
	_ = uc.PubSub.BroadcastLog(ctx, log)
	return &log, nil
}

func (uc *CreateLogUseCase) ExecuteBulk(ctx context.Context, tenantId, userId string, logs []entitylog.Log) ([]entitylog.Log, error) {
	for i := range logs {
		if logs[i].ID == "" {
			logs[i].ID = uuid.New().String()
		}
	}

	if err := uc.TxManager.TransactionExec(ctx, func(txCtx context.Context) error {
		db := uc.TxManager.GetTx(txCtx)
		if err := uc.Repo.CreateBulk(txCtx, db, logs); err != nil {
			return err
		}

		// 2. Create a corresponding task in asyncTask table
		task := &async_task.AsyncTask{
			TaskID:   uuid.New().String(),
			TaskType: async_task.TaskReindex,
			Status:   async_task.StatusPending,
			UserID:   userId,
		}

		if len(tenantId) > 0 {
			task.TenantUID = &tenantId
		}

		if _, err := uc.AsyncTaskRepo.Create(txCtx, db, task); err != nil {
			return err
		}

		// 3. Publish message to SQS
		return uc.QueuePublisher.PublishIndexMessage(txCtx, task.TaskID, logs)
	}); err != nil {
		return nil, err
	}

	// Broadcast logs to redis
	_ = uc.PubSub.BroadcastLogs(ctx, logs)

	return logs, nil

}
