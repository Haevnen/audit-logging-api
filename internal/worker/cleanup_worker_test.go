package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/Haevnen/audit-logging-api/internal/entity/async_task"
	interactorMocks "github.com/Haevnen/audit-logging-api/internal/interactor/mocks"
	repoMocks "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
	"github.com/Haevnen/audit-logging-api/internal/service"
	mockSvc "github.com/Haevnen/audit-logging-api/internal/service/mocks"
	serviceMocks "github.com/Haevnen/audit-logging-api/internal/service/mocks"
	"github.com/Haevnen/audit-logging-api/internal/worker"
	"github.com/Haevnen/audit-logging-api/pkg/utils"
)

func TestCleanUpWorker_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sqs := mockSvc.NewMockSQSPublisher(ctrl)
	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	logRepo := repoMocks.NewMockLogRepository(ctrl)
	tx := interactorMocks.NewMockTxManager(ctrl)
	openSearch := mockSvc.NewMockOpenSearchPublisher(ctrl)

	w := worker.NewCleanUpWorker(sqs, taskRepo, logRepo, tx, openSearch, "cleanup-q")

	before := time.Now()
	msg := service.ReceiveMessage{Message: service.Message{ID: "t1", BeforeDate: utils.Ptr(before)}}

	// task fetch
	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").
		Return(&async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending}, nil).AnyTimes()

	// status updates
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), "t1", async_task.StatusRunning, nil).
		Return(nil).AnyTimes()
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), "t1", async_task.StatusSucceeded, nil).
		Return(nil).AnyTimes()

	// cleanup
	logRepo.EXPECT().CleanupLogsBefore(gomock.Any(), gomock.Any(), gomock.Any(), before).
		Return([]string{"l1"}, nil).AnyTimes()
	openSearch.EXPECT().DeleteLogsBulk(gomock.Any(), []string{"l1"}).Return(nil).AnyTimes()

	// transaction exec
	tx.EXPECT().TransactionExec(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		}).AnyTimes()
	tx.EXPECT().GetTx(gomock.Any()).Return(nil).AnyTimes()

	// sqs behavior
	sqs.EXPECT().ReceiveMessages(gomock.Any(), "cleanup-q", int32(5), int32(20)).
		Return([]service.ReceiveMessage{msg}, nil).AnyTimes()
	sqs.EXPECT().DeleteMessage(gomock.Any(), "cleanup-q", msg.ReceiveHandle).
		Return(nil).AnyTimes()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()
	w.Start(ctx)

	assert.True(t, true)
}

func TestHandleMessage_Success_Cleanup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	logRepo := repoMocks.NewMockLogRepository(ctrl)
	sqs := serviceMocks.NewMockSQSPublisher(ctrl)
	openSearch := serviceMocks.NewMockOpenSearchPublisher(ctrl)
	tx := interactorMocks.NewMockTxManager(ctrl)

	w := worker.NewCleanUpWorker(sqs, taskRepo, logRepo, tx, openSearch, "cleanup-queue")

	before := time.Now()
	task := &async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending, TenantUID: nil, UserID: "u1"}

	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").Return(task, nil)
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusRunning, nil).Return(nil)

	// TransactionExec simulates DB tx
	tx.EXPECT().TransactionExec(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})
	tx.EXPECT().GetTx(gomock.Any()).Return(nil)

	logRepo.EXPECT().CleanupLogsBefore(gomock.Any(), nil, nil, before).Return([]string{"id1", "id2"}, nil)
	openSearch.EXPECT().DeleteLogsBulk(gomock.Any(), []string{"id1", "id2"}).Return(nil)
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusSucceeded, nil).Return(nil)

	msg := service.ReceiveMessage{Message: service.Message{ID: "t1", BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleMessage_TaskFetchErrorCleanUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	w := worker.NewCleanUpWorker(nil, taskRepo, nil, nil, nil, "cleanup")

	before := time.Now()
	taskRepo.EXPECT().GetByID(gomock.Any(), "bad").Return(nil, errors.New("db fail"))

	msg := service.ReceiveMessage{Message: service.Message{ID: "bad", BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
}

func TestHandleMessage_AlreadyProcessed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	w := worker.NewCleanUpWorker(nil, taskRepo, nil, nil, nil, "cleanup")

	before := time.Now()
	task := &async_task.AsyncTask{TaskID: "done", Status: async_task.StatusSucceeded}
	taskRepo.EXPECT().GetByID(gomock.Any(), "done").Return(task, nil)

	msg := service.ReceiveMessage{Message: service.Message{ID: "done", BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleMessage_UpdateStatusFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	w := worker.NewCleanUpWorker(nil, taskRepo, nil, nil, nil, "cleanup")

	before := time.Now()
	task := &async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending}
	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").Return(task, nil)
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusRunning, nil).
		Return(errors.New("update fail"))

	msg := service.ReceiveMessage{Message: service.Message{ID: "t1", BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
}

func TestHandleMessage_CleanupLogsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	logRepo := repoMocks.NewMockLogRepository(ctrl)
	tx := interactorMocks.NewMockTxManager(ctrl)

	w := worker.NewCleanUpWorker(nil, taskRepo, logRepo, tx, nil, "cleanup")

	before := time.Now()
	task := &async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending}
	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").Return(task, nil)
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusRunning, nil).Return(nil)

	tx.EXPECT().TransactionExec(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})
	tx.EXPECT().GetTx(gomock.Any()).Return(nil)

	logRepo.EXPECT().CleanupLogsBefore(gomock.Any(), nil, nil, before).Return(nil, errors.New("cleanup fail"))
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusFailed, gomock.Any()).Return(nil)

	msg := service.ReceiveMessage{Message: service.Message{ID: "t1", BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
}

func TestHandleMessage_OpenSearchDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	logRepo := repoMocks.NewMockLogRepository(ctrl)
	openSearch := serviceMocks.NewMockOpenSearchPublisher(ctrl)
	tx := interactorMocks.NewMockTxManager(ctrl)

	w := worker.NewCleanUpWorker(nil, taskRepo, logRepo, tx, openSearch, "cleanup")

	before := time.Now()
	task := &async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending}
	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").Return(task, nil)
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusRunning, nil).Return(nil)

	tx.EXPECT().TransactionExec(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})
	tx.EXPECT().GetTx(gomock.Any()).Return(nil)

	logRepo.EXPECT().CleanupLogsBefore(gomock.Any(), nil, nil, before).Return([]string{"id1"}, nil)
	openSearch.EXPECT().DeleteLogsBulk(gomock.Any(), []string{"id1"}).Return(errors.New("os fail"))
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusFailed, gomock.Any()).Return(nil)

	msg := service.ReceiveMessage{Message: service.Message{ID: "t1", BeforeDate: &before}}
	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
}
