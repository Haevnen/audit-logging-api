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
	"github.com/google/uuid"
)

type ArchiveWorker struct {
	sqsClient    service.SQSPublisher
	taskRepo     repository.AsyncTaskRepository
	logRepo      repository.LogRepository
	s3Client     service.S3Publisher
	txManager    *interactor.TxManager
	archiveQueue string
}

func NewArchiveWorker(
	sqsClient service.SQSPublisher,
	taskRepo repository.AsyncTaskRepository,
	logRepo repository.LogRepository,
	s3Client service.S3Publisher,
	txManager *interactor.TxManager,
	archiveQueue string,
) *ArchiveWorker {
	return &ArchiveWorker{
		sqsClient:    sqsClient,
		taskRepo:     taskRepo,
		logRepo:      logRepo,
		s3Client:     s3Client,
		txManager:    txManager,
		archiveQueue: archiveQueue,
	}
}

func (w *ArchiveWorker) Start(ctx context.Context) {
	logger := logger.GetLogger()
	for {
		select {
		case <-ctx.Done():
			logger.Info("shutting down archive worker")
			return
		default:
			// Long poll for archival tasks
			msgs, err := w.sqsClient.ReceiveMessages(ctx, w.archiveQueue, 5, 20)
			if err != nil {
				logger.Warning("failed to receive", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, m := range msgs {
				if err := w.handleMessage(ctx, m); err != nil {
					logger.Warning("failed to handle message", err)
				}
				// Always delete to avoid retries storm
				_ = w.sqsClient.DeleteMessage(ctx, w.archiveQueue, m.ReceiveHandle)
			}
		}
	}
}

func (w *ArchiveWorker) handleMessage(ctx context.Context, msg service.ReceiveMessage) error {
	logger := logger.GetLogger()
	taskId, beforeDate := msg.Message.ID, msg.Message.BeforeDate
	logger.WithFields(map[string]interface{}{
		"taskId":     taskId,
		"beforeDate": beforeDate,
	}).Info("received message")

	// Load async task
	task, err := w.taskRepo.GetByID(ctx, taskId)
	if err != nil {
		return fmt.Errorf("task fetch failed: %w", err)
	}

	if task.Status != async_task.StatusPending {
		logger.WithFields(map[string]interface{}{
			"taskId": taskId,
			"status": task.Status,
		}).Info("Already processed")
		return nil
	}

	// Update -> RUNNING
	if err := w.taskRepo.UpdateStatus(ctx, nil, taskId, async_task.StatusRunning, nil); err != nil {
		logger.WithField("error", err).Error("status update failed")
		return fmt.Errorf("status update failed: %w", err)
	}

	// Query logs to archive (your repo should accept filters from task.Payload)
	logs, err := w.logRepo.FindLogsForArchival(ctx, task.TenantUID, *beforeDate)
	if err != nil {
		_ = w.taskRepo.UpdateStatus(ctx, nil, taskId, async_task.StatusFailed, utils.Ptr(err.Error()))
		return fmt.Errorf("log query failed: %w", err)
	}

	// Write logs to S3
	logger.Info("Uploading logs to S3")
	if err := w.s3Client.UploadLogs(ctx, taskId, logs); err != nil {
		_ = w.taskRepo.UpdateStatus(ctx, nil, taskId, async_task.StatusFailed, utils.Ptr(err.Error()))
		return fmt.Errorf("s3 upload failed: %w", err)
	}
	logger.Info("Uploaded logs to S3")

	// Start a transaction to update status to success and publish cleanup message
	if err := w.txManager.TransactionExec(ctx, func(txCtx context.Context) error {
		db := w.txManager.GetTx(txCtx)

		// Update → SUCCESS
		logger.Info("Updating status to success")
		if err := w.taskRepo.UpdateStatus(ctx, db, taskId, async_task.StatusSucceeded, nil); err != nil {
			return fmt.Errorf("final status update failed: %w", err)
		}

		// Publish message to cleanup queue
		newTask := &async_task.AsyncTask{
			TaskID:   uuid.New().String(),
			TaskType: async_task.TaskLogCleanup,
			Status:   async_task.StatusPending,
			UserID:   task.UserID,
		}

		if task.TenantUID != nil && len(*task.TenantUID) > 0 {
			newTask.TenantUID = task.TenantUID
		}

		logger.Info("Publishing cleanup message")
		newCreatedTask, err := w.taskRepo.Create(txCtx, db, newTask)
		if err != nil {
			return err
		}

		return w.sqsClient.PublishCleanUpMessage(txCtx, newCreatedTask.TaskID, *beforeDate)
	}); err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	logger.Info("Published cleanup message")
	return nil
}
