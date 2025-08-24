package tenant_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	entitytenant "github.com/Haevnen/audit-logging-api/internal/entity/tenant"
	uc "github.com/Haevnen/audit-logging-api/internal/usecase/tenant"

	repoMocks "github.com/Haevnen/audit-logging-api/internal/repository/mocks"
)

func TestCreateTenantUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockTenantRepository(ctrl)
	ctx := context.Background()

	expected := &entitytenant.Tenant{ID: "tenant-123", Name: "TestTenant"}

	mockRepo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, t *entitytenant.Tenant) (*entitytenant.Tenant, error) {
			// Simulate DB filling in ID
			t.ID = expected.ID
			return t, nil
		})

	ucase := uc.NewCreateTenantUseCase(mockRepo)

	result, err := ucase.Execute(ctx, "TestTenant")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Name, result.Name)
}

func TestCreateTenantUseCase_Execute_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockTenantRepository(ctrl)
	ctx := context.Background()

	mockRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(nil, assert.AnError)

	ucase := uc.NewCreateTenantUseCase(mockRepo)

	result, err := ucase.Execute(ctx, "BadTenant")
	assert.Error(t, err)
	assert.Nil(t, result)
}
