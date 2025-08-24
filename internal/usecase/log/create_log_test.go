package log_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/Haevnen/audit-logging-api/internal/entity/async_task"
	entitylog "github.com/Haevnen/audit-logging-api/internal/entity/log"
	uc "github.com/Haevnen/audit-logging-api/internal/usecase/log"

	intMocks "github.com/Haevnen/audit-logging-api/internal/interactor/mocks"
	repoMocks "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
	svcMocks "github.com/Haevnen/audit-logging-api/internal/service/mocks"
)

func TestCreateLogUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogRepository(ctrl)
	mockAsync := repoMocks.NewMockAsyncTaskRepository(ctrl)
	mockTx := intMocks.NewMockTxManager(ctrl)
	mockSQS := svcMocks.NewMockSQSPublisher(ctrl)
	mockPub := svcMocks.NewMockPubSub(ctrl)

	ctx := context.Background()
	logEntry := entitylog.Log{Message: "test"}

	mockTx.EXPECT().
		TransactionExec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})
	mockTx.EXPECT().GetTx(gomock.Any()).Return(&gorm.DB{})

	mockRepo.EXPECT().CreateBulk(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockAsync.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&async_task.AsyncTask{TaskID: uuid.New().String()}, nil)
	mockSQS.EXPECT().PublishIndexMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockPub.EXPECT().BroadcastLog(gomock.Any(), gomock.Any()).Return(nil)

	ucase := uc.NewCreateLogUseCase(mockRepo, mockTx, mockSQS, mockPub, mockAsync)

	result, err := ucase.Execute(ctx, "tenant-1", "user-1", logEntry)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
}

func TestCreateLogUseCase_Execute_Fail_Repo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogRepository(ctrl)
	mockAsync := repoMocks.NewMockAsyncTaskRepository(ctrl)
	mockTx := intMocks.NewMockTxManager(ctrl)
	mockSQS := svcMocks.NewMockSQSPublisher(ctrl)
	mockPub := svcMocks.NewMockPubSub(ctrl)

	ctx := context.Background()
	logEntry := entitylog.Log{Message: "fail repo"}

	mockTx.EXPECT().
		TransactionExec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})
	mockTx.EXPECT().GetTx(gomock.Any()).Return(&gorm.DB{})

	mockRepo.EXPECT().CreateBulk(gomock.Any(), gomock.Any(), gomock.Any()).Return(assert.AnError)

	ucase := uc.NewCreateLogUseCase(mockRepo, mockTx, mockSQS, mockPub, mockAsync)

	result, err := ucase.Execute(ctx, "tenant-1", "user-1", logEntry)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCreateLogUseCase_Execute_Fail_AsyncTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogRepository(ctrl)
	mockAsync := repoMocks.NewMockAsyncTaskRepository(ctrl)
	mockTx := intMocks.NewMockTxManager(ctrl)
	mockSQS := svcMocks.NewMockSQSPublisher(ctrl)
	mockPub := svcMocks.NewMockPubSub(ctrl)

	ctx := context.Background()
	logEntry := entitylog.Log{Message: "fail async"}

	mockTx.EXPECT().
		TransactionExec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})
	mockTx.EXPECT().GetTx(gomock.Any()).Return(&gorm.DB{})

	mockRepo.EXPECT().CreateBulk(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockAsync.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, assert.AnError)

	ucase := uc.NewCreateLogUseCase(mockRepo, mockTx, mockSQS, mockPub, mockAsync)

	result, err := ucase.Execute(ctx, "tenant-1", "user-1", logEntry)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCreateLogUseCase_Execute_Fail_SQS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogRepository(ctrl)
	mockAsync := repoMocks.NewMockAsyncTaskRepository(ctrl)
	mockTx := intMocks.NewMockTxManager(ctrl)
	mockSQS := svcMocks.NewMockSQSPublisher(ctrl)
	mockPub := svcMocks.NewMockPubSub(ctrl)

	ctx := context.Background()
	logEntry := entitylog.Log{Message: "fail sqs"}

	mockTx.EXPECT().
		TransactionExec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})
	mockTx.EXPECT().GetTx(gomock.Any()).Return(&gorm.DB{})

	mockRepo.EXPECT().CreateBulk(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockAsync.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&async_task.AsyncTask{TaskID: uuid.New().String()}, nil)
	mockSQS.EXPECT().PublishIndexMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(assert.AnError)

	ucase := uc.NewCreateLogUseCase(mockRepo, mockTx, mockSQS, mockPub, mockAsync)

	result, err := ucase.Execute(ctx, "tenant-1", "user-1", logEntry)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCreateLogUseCase_ExecuteBulk_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogRepository(ctrl)
	mockAsync := repoMocks.NewMockAsyncTaskRepository(ctrl)
	mockTx := intMocks.NewMockTxManager(ctrl)
	mockSQS := svcMocks.NewMockSQSPublisher(ctrl)
	mockPub := svcMocks.NewMockPubSub(ctrl)

	ctx := context.Background()
	logs := []entitylog.Log{{Message: "bulk1"}, {Message: "bulk2"}}

	mockTx.EXPECT().
		TransactionExec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})
	mockTx.EXPECT().GetTx(gomock.Any()).Return(&gorm.DB{})

	mockRepo.EXPECT().CreateBulk(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockAsync.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&async_task.AsyncTask{TaskID: uuid.New().String()}, nil)
	mockSQS.EXPECT().PublishIndexMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockPub.EXPECT().BroadcastLogs(gomock.Any(), gomock.Any()).Return(nil)

	ucase := uc.NewCreateLogUseCase(mockRepo, mockTx, mockSQS, mockPub, mockAsync)

	result, err := ucase.ExecuteBulk(ctx, "tenant-1", "user-1", logs)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.NotEmpty(t, result[0].ID)
}
