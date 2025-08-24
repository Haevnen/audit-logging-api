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
	repoMocks "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
	"github.com/Haevnen/audit-logging-api/internal/service"
	mockSvc "github.com/Haevnen/audit-logging-api/internal/service/mocks"
	serviceMocks "github.com/Haevnen/audit-logging-api/internal/service/mocks"
	"github.com/Haevnen/audit-logging-api/internal/worker"
)

func TestIndexWorker_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sqs := mockSvc.NewMockSQSPublisher(ctrl)
	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	openSearch := mockSvc.NewMockOpenSearchPublisher(ctrl)
	tx := interactorMocks.NewMockTxManager(ctrl)

	w := worker.NewIndexWorker(sqs, tx, taskRepo, openSearch, "index-q")

	logs := []log.Log{{ID: "l1"}}
	msg := service.ReceiveMessage{Message: service.Message{ID: "t1", Logs: &logs}}

	// task fetch
	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").
		Return(&async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending}, nil).AnyTimes()

	// status updates
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), "t1", async_task.StatusRunning, nil).
		Return(nil).AnyTimes()
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), "t1", async_task.StatusSucceeded, nil).
		Return(nil).AnyTimes()

	// openSearch indexing
	openSearch.EXPECT().IndexLogsBulk(gomock.Any(), logs).Return(nil).AnyTimes()

	// transaction exec
	tx.EXPECT().TransactionExec(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		}).AnyTimes()
	tx.EXPECT().GetTx(gomock.Any()).Return(nil).AnyTimes()

	// sqs behavior
	sqs.EXPECT().ReceiveMessages(gomock.Any(), "index-q", int32(5), int32(20)).
		Return([]service.ReceiveMessage{msg}, nil).AnyTimes()
	sqs.EXPECT().DeleteMessage(gomock.Any(), "index-q", msg.ReceiveHandle).
		Return(nil).AnyTimes()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()
	w.Start(ctx)

	assert.True(t, true)
}

func TestHandleMessage_Success_Index(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	openSearch := serviceMocks.NewMockOpenSearchPublisher(ctrl)
	tx := interactorMocks.NewMockTxManager(ctrl)

	w := worker.NewIndexWorker(nil, tx, taskRepo, openSearch, "index-q")

	task := &async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending}
	msgLogs := []log.Log{{ID: "l1"}}
	msg := service.ReceiveMessage{Message: service.Message{ID: "t1", Logs: &msgLogs}}

	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").Return(task, nil)
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusRunning, nil).Return(nil)

	// simulate tx success
	tx.EXPECT().TransactionExec(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})
	tx.EXPECT().GetTx(gomock.Any()).Return(nil)

	openSearch.EXPECT().IndexLogsBulk(gomock.Any(), msgLogs).Return(nil)
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusSucceeded, nil).Return(nil)

	err := w.HandleMessage(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleMessage_TaskFetchError_Index(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	w := worker.NewIndexWorker(nil, nil, taskRepo, nil, "index-q")

	taskRepo.EXPECT().GetByID(gomock.Any(), "bad").Return(nil, errors.New("db fail"))

	msg := service.ReceiveMessage{Message: service.Message{ID: "bad", Logs: &[]log.Log{}}}
	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
}

func TestHandleMessage_AlreadyProcessed_Index(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	w := worker.NewIndexWorker(nil, nil, taskRepo, nil, "index-q")

	task := &async_task.AsyncTask{TaskID: "done", Status: async_task.StatusSucceeded}
	taskRepo.EXPECT().GetByID(gomock.Any(), "done").Return(task, nil)

	msg := service.ReceiveMessage{Message: service.Message{ID: "done", Logs: &[]log.Log{}}}
	err := w.HandleMessage(context.Background(), msg)
	assert.NoError(t, err)
}

func TestHandleMessage_UpdateStatusRunningFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	w := worker.NewIndexWorker(nil, nil, taskRepo, nil, "index-q")

	task := &async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending}
	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").Return(task, nil)
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusRunning, nil).
		Return(errors.New("update fail"))

	msg := service.ReceiveMessage{Message: service.Message{ID: "t1", Logs: &[]log.Log{}}}
	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
}

func TestHandleMessage_IndexLogsBulkError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskRepo := repoMocks.NewMockAsyncTaskRepository(ctrl)
	openSearch := serviceMocks.NewMockOpenSearchPublisher(ctrl)
	tx := interactorMocks.NewMockTxManager(ctrl)

	w := worker.NewIndexWorker(nil, tx, taskRepo, openSearch, "index-q")

	task := &async_task.AsyncTask{TaskID: "t1", Status: async_task.StatusPending}
	msgLogs := []log.Log{{ID: "l1"}}
	msg := service.ReceiveMessage{Message: service.Message{ID: "t1", Logs: &msgLogs}}

	taskRepo.EXPECT().GetByID(gomock.Any(), "t1").Return(task, nil)
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusRunning, nil).Return(nil)

	tx.EXPECT().TransactionExec(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})
	tx.EXPECT().GetTx(gomock.Any()).Return(nil)

	openSearch.EXPECT().IndexLogsBulk(gomock.Any(), msgLogs).Return(errors.New("os fail"))
	taskRepo.EXPECT().UpdateStatus(gomock.Any(), nil, "t1", async_task.StatusFailed, gomock.Any()).Return(nil)

	err := w.HandleMessage(context.Background(), msg)
	assert.Error(t, err)
}
