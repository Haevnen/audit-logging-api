package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/Haevnen/audit-logging-api/internal/entity/async_task"
	"github.com/Haevnen/audit-logging-api/internal/interactor"
	"github.com/Haevnen/audit-logging-api/internal/repository"
	"github.com/Haevnen/audit-logging-api/internal/service"
	"github.com/Haevnen/audit-logging-api/pkg/logger"
	"github.com/Haevnen/audit-logging-api/pkg/utils"
)

type CleanUpWorker struct {
	sqsClient    service.SQSPublisher
	taskRepo     repository.AsyncTaskRepository
	logRepo      repository.LogRepository
	txManager    interactor.TxManager
	openSearch   service.OpenSearchPublisher
	cleanupQueue string
}

func NewCleanUpWorker(
	sqsClient service.SQSPublisher,
	taskRepo repository.AsyncTaskRepository,
	logRepo repository.LogRepository,
	txManager interactor.TxManager,
	openSearch service.OpenSearchPublisher,
	cleanupQueue string,
) *CleanUpWorker {
	return &CleanUpWorker{
		sqsClient:    sqsClient,
		taskRepo:     taskRepo,
		logRepo:      logRepo,
		txManager:    txManager,
		openSearch:   openSearch,
		cleanupQueue: cleanupQueue,
	}
}

func (w *CleanUpWorker) Start(ctx context.Context) {
	logger := logger.GetLogger()
	for {
		select {
		case <-ctx.Done():
			logger.Info("shutting down cleanup worker")
			return
		default:
			msgs, err := w.sqsClient.ReceiveMessages(ctx, w.cleanupQueue, 5, 20)
			if err != nil {
				logger.Warning("failed to receive cleanup messages", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, m := range msgs {
				if err := w.HandleMessage(ctx, m); err != nil {
					logger.Warning("failed to handle cleanup message", err)
				}
				// Always delete to avoid retries storm
				_ = w.sqsClient.DeleteMessage(ctx, w.cleanupQueue, m.ReceiveHandle)
			}
		}
	}
}

func (w *CleanUpWorker) HandleMessage(ctx context.Context, msg service.ReceiveMessage) error {
	log := logger.GetLogger()
	taskId, beforeDate := msg.Message.ID, msg.Message.BeforeDate
	log.WithFields(map[string]interface{}{
		"taskId":     taskId,
		"beforeDate": beforeDate,
	}).Info("cleanup worker received message")

	task, err := w.taskRepo.GetByID(ctx, taskId)
	if err != nil {
		return fmt.Errorf("task fetch failed: %w", err)
	}

	if task.Status != async_task.StatusPending {
		log.WithField("taskId", taskId).Info("cleanup task already processed")
		return nil
	}

	// Update status → RUNNING
	if err := w.taskRepo.UpdateStatus(ctx, nil, taskId, async_task.StatusRunning, nil); err != nil {
		return fmt.Errorf("status update failed: %w", err)
	}

	// Perform cleanup
	if err := w.txManager.TransactionExec(ctx, func(txCtx context.Context) error {
		db := w.txManager.GetTx(txCtx)

		// Delete logs
		ids, err := w.logRepo.CleanupLogsBefore(txCtx, db, task.TenantUID, *beforeDate)
		if err != nil {
			return err
		}

		if len(ids) > 0 {
			if err := w.openSearch.DeleteLogsBulk(txCtx, ids); err != nil {
				return err
			}
		}

		// Update status → COMPLETED
		return w.taskRepo.UpdateStatus(txCtx, db, taskId, async_task.StatusSucceeded, nil)
	}); err != nil {
		_ = w.taskRepo.UpdateStatus(ctx, nil, taskId, async_task.StatusFailed, utils.Ptr(err.Error()))
		return fmt.Errorf("cleanup failed: %w", err)
	}

	log.WithField("taskId", taskId).Info("cleanup succeeded")
	return nil
}
