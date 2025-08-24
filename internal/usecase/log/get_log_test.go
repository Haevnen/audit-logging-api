package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	entitylog "github.com/Haevnen/audit-logging-api/internal/entity/log"
	uc "github.com/Haevnen/audit-logging-api/internal/usecase/log"

	repoMocks "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
)

func TestGetLogUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogRepository(ctrl)

	ctx := context.Background()
	expected := &entitylog.Log{ID: "log-123", TenantID: "tenant-1"}

	mockRepo.EXPECT().
		GetByID(ctx, "log-123", "tenant-1").
		Return(expected, nil)

	ucase := uc.NewGetLogUseCase(mockRepo)

	result, err := ucase.Execute(ctx, "log-123", "tenant-1")
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetLogUseCase_Execute_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogRepository(ctrl)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetByID(ctx, "bad-id", "tenant-1").
		Return(nil, assert.AnError)

	ucase := uc.NewGetLogUseCase(mockRepo)

	result, err := ucase.Execute(ctx, "bad-id", "tenant-1")
	assert.Error(t, err)
	assert.Nil(t, result)
}
