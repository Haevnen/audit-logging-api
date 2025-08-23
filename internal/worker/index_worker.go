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

type IndexWorker struct {
	sqsClient  service.SQSPublisher
	txManager  *interactor.TxManager
	taskRepo   repository.AsyncTaskRepository
	openSearch service.OpenSearchPublisher
	indexQueue string
}

func NewIndexWorker(
	sqsClient service.SQSPublisher,
	txManager *interactor.TxManager,
	taskRepo repository.AsyncTaskRepository,
	openSearch service.OpenSearchPublisher,
	indexQueue string,
) *IndexWorker {
	return &IndexWorker{
		sqsClient:  sqsClient,
		taskRepo:   taskRepo,
		openSearch: openSearch,
		indexQueue: indexQueue,
		txManager:  txManager,
	}
}

func (w *IndexWorker) Start(ctx context.Context) {
	logger := logger.GetLogger()
	for {
		select {
		case <-ctx.Done():
			logger.Info("shutting down index worker")
			return
		default:
			msgs, err := w.sqsClient.ReceiveMessages(ctx, w.indexQueue, 5, 20)
			if err != nil {
				logger.Warning("failed to receive index messages", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, m := range msgs {
				if err := w.handleMessage(ctx, m); err != nil {
					logger.Warning("failed to handle index message", err)
				}
				// Always delete to avoid retries storm
				_ = w.sqsClient.DeleteMessage(ctx, w.indexQueue, m.ReceiveHandle)
			}
		}
	}
}

func (w *IndexWorker) handleMessage(ctx context.Context, msg service.ReceiveMessage) error {
	log := logger.GetLogger()
	taskId := msg.Message.ID
	log.WithFields(map[string]interface{}{
		"taskId": taskId,
	}).Info("Index worker received message")

	task, err := w.taskRepo.GetByID(ctx, taskId)
	if err != nil {
		return fmt.Errorf("task fetch failed: %w", err)
	}

	if task.Status != async_task.StatusPending {
		log.WithField("taskId", taskId).Info("index task already processed")
		return nil
	}

	// Update status → RUNNING
	if err := w.taskRepo.UpdateStatus(ctx, nil, taskId, async_task.StatusRunning, nil); err != nil {
		return fmt.Errorf("status update failed: %w", err)
	}

	// Perform indexing
	if err := w.txManager.TransactionExec(ctx, func(txCtx context.Context) error {
		db := w.txManager.GetTx(txCtx)

		// Sync with OpenSearch
		if err := w.openSearch.IndexLogsBulk(context.Background(), *msg.Message.Logs); err != nil {
			logger.GetLogger().Errorf("failed to index log to opensearch: %v", err)
			return err
		}

		// Update status → COMPLETED
		return w.taskRepo.UpdateStatus(txCtx, db, taskId, async_task.StatusSucceeded, nil)
	}); err != nil {
		_ = w.taskRepo.UpdateStatus(ctx, nil, taskId, async_task.StatusFailed, utils.Ptr(err.Error()))
		return fmt.Errorf("indexing failed: %w", err)
	}

	log.WithField("taskId", taskId).Info("index succeeded")
	return nil
}
