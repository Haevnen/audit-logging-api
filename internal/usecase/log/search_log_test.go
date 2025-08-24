package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	entitylog "github.com/Haevnen/audit-logging-api/internal/entity/log"
	uc "github.com/Haevnen/audit-logging-api/internal/usecase/log"

	"github.com/Haevnen/audit-logging-api/internal/repository"
	repoMocks "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
)

func TestSearchLogsUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogSearchRepository(ctrl)
	ctx := context.Background()
	filters := repository.LogSearchFilters{}

	expected := &repository.SearchResult{Total: 1}

	mockRepo.EXPECT().
		Search(ctx, filters).
		Return(expected, nil)

	ucase := uc.NewSearchLogsUseCase(mockRepo)

	result, err := ucase.Execute(ctx, filters)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSearchLogsUseCase_Execute_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogSearchRepository(ctrl)
	ctx := context.Background()
	filters := repository.LogSearchFilters{}

	mockRepo.EXPECT().
		Search(ctx, filters).
		Return(nil, assert.AnError)

	ucase := uc.NewSearchLogsUseCase(mockRepo)

	result, err := ucase.Execute(ctx, filters)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSearchLogsUseCase_Stream_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogSearchRepository(ctrl)
	ctx := context.Background()
	filters := repository.LogSearchFilters{}

	mockRepo.EXPECT().
		Stream(ctx, filters, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ repository.LogSearchFilters, fn func(entitylog.Log) error) error {
			return fn(entitylog.Log{ID: "log-1"})
		})

	ucase := uc.NewSearchLogsUseCase(mockRepo)

	err := ucase.Stream(ctx, filters, func(l entitylog.Log) error {
		assert.Equal(t, "log-1", l.ID)
		return nil
	})
	assert.NoError(t, err)
}

func TestSearchLogsUseCase_Stream_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockLogSearchRepository(ctrl)
	ctx := context.Background()
	filters := repository.LogSearchFilters{}

	mockRepo.EXPECT().
		Stream(ctx, filters, gomock.Any()).
		Return(assert.AnError)

	ucase := uc.NewSearchLogsUseCase(mockRepo)

	err := ucase.Stream(ctx, filters, func(l entitylog.Log) error {
		return nil
	})
	assert.Error(t, err)
}
