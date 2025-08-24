package log_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	entitylog "github.com/Haevnen/audit-logging-api/internal/entity/log"
	uc "github.com/Haevnen/audit-logging-api/internal/usecase/log"

	repoMocks "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
)

func TestGetStatsUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogRepository(ctrl)
	ctx := context.Background()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	expected := []entitylog.LogStats{
		{
			Day:          start,
			Total:        7,
			ActionCreate: 5,
			ActionDelete: 2,
		},
	}

	mockRepo.EXPECT().
		GetStats(ctx, "tenant-1", start, end).
		Return(expected, nil)

	ucase := uc.NewGetStatsUseCase(mockRepo)

	result, err := ucase.Execute(ctx, "tenant-1", start, end)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetStatsUseCase_Execute_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogRepository(ctrl)
	ctx := context.Background()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	mockRepo.EXPECT().
		GetStats(ctx, "tenant-1", start, end).
		Return(nil, assert.AnError)

	ucase := uc.NewGetStatsUseCase(mockRepo)

	result, err := ucase.Execute(ctx, "tenant-1", start, end)
	assert.Error(t, err)
	assert.Nil(t, result)
}
