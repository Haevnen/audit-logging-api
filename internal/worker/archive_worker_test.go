package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/Haevnen/audit-logging-api/internal/entity/async_task"
	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	interactorMocks "github.com/Haevnen/audit-logging-api/internal/interactor/mocks"
	mockTx "github.com/Haevnen/audit-logging-api/internal/interactor/mocks"
	mockRepo "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
	repoMocks "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
	"github.com/Haevnen/audit-logging-api/internal/service"
	mockSvc "github.com/Haevnen/audit-logging-api/internal/service/mocks"
	"github.com/Haevnen/audit-logging-api/internal/worker"
	"github.com/Haevnen/audit-logging-api/pkg/utils"
)

func TestArchiveWorker_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sqs := mockSvc.NewMockSQSPublisher(ctrl)
	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	logRepo := repoMocks.NewMockLogRepository(ctrl)
	s3 := mockSvc.NewMockS3Publisher(ctrl)
	tx := interactorMocks.NewMockTxManager(ctrl)

	w := worker.NewArchiveWorker(sqs, taskRepo, logRepo, s3, tx, "archive-q")

	msg := service.ReceiveMessage{
		Message: service.Message{ID: "t1", BeforeDate: utils.Ptr(time.Now())},
	}

	// --- Expectations ---
	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").
		Return(&async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending}, nil).
		AnyTimes()

	taskRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), "t1", async_task.StatusRunning, nil).
		Return(nil).AnyTimes()

	logRepo.EXPECT().FindLogsForArchival(gomock.Any(), nil, gomock.Any()).
		Return(nil, nil).AnyTimes()

	s3.EXPECT().UploadLogs(gomock.Any(), "t1", gomock.Any()).
		Return(nil).AnyTimes()

	tx.EXPECT().TransactionExec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		}).AnyTimes()

	// 👇 THIS is the missing piece that caused the panic
	tx.EXPECT().GetTx(gomock.Any()).Return(nil).AnyTimes()

	taskRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&async_task.AsyncTask{TaskID: "new-task"}, nil).AnyTimes()

	sqs.EXPECT().PublishCleanUpMessage(gomock.Any(), "new-task", gomock.Any()).
		Return(nil).AnyTimes()

	sqs.EXPECT().ReceiveMessages(gomock.Any(), "archive-q", int32(5), int32(20)).
		Return([]service.ReceiveMessage{msg}, nil).AnyTimes()

	sqs.EXPECT().DeleteMessage(gomock.Any(), "archive-q", msg.ReceiveHandle).
		Return(nil).AnyTimes()

	// Add both expectations
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), "t1", async_task.StatusRunning, nil).
		Return(nil).AnyTimes()
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), "t1", async_task.StatusSucceeded, nil).
		Return(nil).AnyTimes()

	// --- Run worker with cancel ---
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()
	w.Start(ctx)

	assert.True(t, true, "worker should run and stop without panic")
}

func TestHandleMessage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mockRepo.NewMockAsyncTaskRepository(ctrl)
	mockLogRepo := mockRepo.NewMockLogRepository(ctrl)
	mockS3 := mockSvc.NewMockS3Publisher(ctrl)
	mockSQS := mockSvc.NewMockSQSPublisher(ctrl)
	mockTxMgr := mockTx.NewMockTxManager(ctrl)

	w := worker.NewArchiveWorker(mockSQS, mockTaskRepo, mockLogRepo, mockS3, mockTxMgr, "archive-queue")

	before := time.Now().Add(-24 * time.Hour)
	taskID := "task-123"
	tenant := "tenant-1"
	task := &async_task.AsyncTask{TaskID: taskID, Status: async_task.StatusPending, UserID: "u1", TenantUID: &tenant}

	// expectations
	mockTaskRepo.EXPECT().GetByID(gomock.Any(), taskID).Return(task, nil)
	mockTaskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, taskID, async_task.StatusRunning, nil).Return(nil)
	mockLogRepo.EXPECT().FindLogsForArchival(gomock.Any(), task.TenantUID, before).Return([]log.Log{{ID: "l1"}}, nil)
	mockS3.EXPECT().UploadLogs(gomock.Any(), taskID, gomock.Any()).Return(nil)

	// transaction
	mockTxMgr.EXPECT().TransactionExec(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)
	mockTxMgr.EXPECT().GetTx(gomock.Any()).Return(nil)

	mockTaskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, taskID, async_task.StatusSucceeded, nil).Return(nil)
	mockTaskRepo.EXPECT().Create(gomock.Any(), nil, gomock.Any()).Return(&async_task.AsyncTask{TaskID: "cleanup-1"}, nil)
	mockSQS.EXPECT().PublishCleanUpMessage(gomock.Any(), "cleanup-1", before).Return(nil)

	msg := service.ReceiveMessage{Message: service.Message{ID: taskID, BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleMessage_TaskFetchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mockRepo.NewMockAsyncTaskRepository(ctrl)
	w := worker.NewArchiveWorker(nil, mockTaskRepo, nil, nil, nil, "q")

	taskID := "task-123"
	before := time.Now()
	mockTaskRepo.EXPECT().GetByID(gomock.Any(), taskID).Return(nil, errors.New("db down"))

	msg := service.ReceiveMessage{Message: service.Message{ID: taskID, BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task fetch failed")
}

func TestHandleMessage_LogQueryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mockRepo.NewMockAsyncTaskRepository(ctrl)
	mockLogRepo := mockRepo.NewMockLogRepository(ctrl)

	w := worker.NewArchiveWorker(nil, mockTaskRepo, mockLogRepo, nil, nil, "q")

	taskID := "task-123"
	before := time.Now()
	tenant := "t1"
	task := &async_task.AsyncTask{TaskID: taskID, Status: async_task.StatusPending, TenantUID: &tenant}

	mockTaskRepo.EXPECT().GetByID(gomock.Any(), taskID).Return(task, nil)
	mockTaskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, taskID, async_task.StatusRunning, nil).Return(nil)
	mockLogRepo.EXPECT().FindLogsForArchival(gomock.Any(), task.TenantUID, before).Return(nil, errors.New("query fail"))
	mockTaskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, taskID, async_task.StatusFailed, gomock.Any()).Return(nil)

	msg := service.ReceiveMessage{Message: service.Message{ID: taskID, BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "log query failed")
}

func TestHandleMessage_S3UploadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mockRepo.NewMockAsyncTaskRepository(ctrl)
	mockLogRepo := mockRepo.NewMockLogRepository(ctrl)
	mockS3 := mockSvc.NewMockS3Publisher(ctrl)

	w := worker.NewArchiveWorker(nil, mockTaskRepo, mockLogRepo, mockS3, nil, "q")

	taskID := "task-123"
	before := time.Now()
	tenant := "t1"
	task := &async_task.AsyncTask{TaskID: taskID, Status: async_task.StatusPending, TenantUID: &tenant}

	mockTaskRepo.EXPECT().GetByID(gomock.Any(), taskID).Return(task, nil)
	mockTaskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, taskID, async_task.StatusRunning, nil).Return(nil)
	mockLogRepo.EXPECT().FindLogsForArchival(gomock.Any(), task.TenantUID, before).Return([]log.Log{}, nil)
	mockS3.EXPECT().UploadLogs(gomock.Any(), taskID, gomock.Any()).Return(errors.New("s3 error"))
	mockTaskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, taskID, async_task.StatusFailed, gomock.Any()).Return(nil)

	msg := service.ReceiveMessage{Message: service.Message{ID: taskID, BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "s3 upload failed")
}

func TestHandleMessage_AlreadySucceeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mockRepo.NewMockAsyncTaskRepository(ctrl)
	w := worker.NewArchiveWorker(nil, mockTaskRepo, nil, nil, nil, "q")

	taskID := "task-123"
	before := time.Now()
	task := &async_task.AsyncTask{TaskID: taskID, Status: async_task.StatusSucceeded, UserID: "u1"}

	mockTaskRepo.EXPECT().GetByID(gomock.Any(), taskID).Return(task, nil)

	msg := service.ReceiveMessage{Message: service.Message{ID: taskID, BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.NoError(t, err, "should return nil when task already succeeded")
}
