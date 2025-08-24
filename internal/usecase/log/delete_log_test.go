package log_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/Haevnen/audit-logging-api/internal/entity/async_task"
	uc "github.com/Haevnen/audit-logging-api/internal/usecase/log"

	intMocks "github.com/Haevnen/audit-logging-api/internal/interactor/mocks"
	repoMocks "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
	svcMocks "github.com/Haevnen/audit-logging-api/internal/service/mocks"
)

func TestDeleteLogUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAsync := repoMocks.NewMockAsyncTaskRepository(ctrl)
	mockTx := intMocks.NewMockTxManager(ctrl)
	mockSQS := svcMocks.NewMockSQSPublisher(ctrl)

	ctx := context.Background()
	beforeDate := time.Now().Add(-24 * time.Hour)

	mockTx.EXPECT().
		TransactionExec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})
	mockTx.EXPECT().GetTx(gomock.Any()).Return(&gorm.DB{})

	mockAsync.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&async_task.AsyncTask{TaskID: uuid.New().String()}, nil)

	mockSQS.EXPECT().
		PublishArchiveMessage(gomock.Any(), gomock.Any(), beforeDate).
		Return(nil)

	ucase := uc.NewDeleteLogUseCase(mockAsync, mockSQS, mockTx)

	err := ucase.Execute(ctx, "tenant-1", "user-1", beforeDate)
	assert.NoError(t, err)
}

func TestDeleteLogUseCase_Execute_Fail_AsyncTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAsync := repoMocks.NewMockAsyncTaskRepository(ctrl)
	mockTx := intMocks.NewMockTxManager(ctrl)
	mockSQS := svcMocks.NewMockSQSPublisher(ctrl)

	ctx := context.Background()
	beforeDate := time.Now().Add(-24 * time.Hour)

	mockTx.EXPECT().
		TransactionExec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})
	mockTx.EXPECT().GetTx(gomock.Any()).Return(&gorm.DB{})

	mockAsync.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	ucase := uc.NewDeleteLogUseCase(mockAsync, mockSQS, mockTx)

	err := ucase.Execute(ctx, "tenant-1", "user-1", beforeDate)
	assert.Error(t, err)
}

func TestDeleteLogUseCase_Execute_Fail_SQS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAsync := repoMocks.NewMockAsyncTaskRepository(ctrl)
	mockTx := intMocks.NewMockTxManager(ctrl)
	mockSQS := svcMocks.NewMockSQSPublisher(ctrl)

	ctx := context.Background()
	beforeDate := time.Now().Add(-24 * time.Hour)

	mockTx.EXPECT().
		TransactionExec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})
	mockTx.EXPECT().GetTx(gomock.Any()).Return(&gorm.DB{})

	mockAsync.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&async_task.AsyncTask{TaskID: uuid.New().String()}, nil)

	mockSQS.EXPECT().
		PublishArchiveMessage(gomock.Any(), gomock.Any(), beforeDate).
		Return(assert.AnError)

	ucase := uc.NewDeleteLogUseCase(mockAsync, mockSQS, mockTx)

	err := ucase.Execute(ctx, "tenant-1", "user-1", beforeDate)
	assert.Error(t, err)
}
