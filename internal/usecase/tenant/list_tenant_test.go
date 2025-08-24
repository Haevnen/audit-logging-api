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

func TestListTenantsUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockTenantRepository(ctrl)
	ctx := context.Background()

	expected := []entitytenant.Tenant{
		{ID: "t1", Name: "Tenant One"},
		{ID: "t2", Name: "Tenant Two"},
	}

	mockRepo.EXPECT().
		List(ctx).
		Return(expected, nil)

	ucase := uc.NewListTenantsUseCase(mockRepo)

	result, err := ucase.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestListTenantsUseCase_Execute_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockTenantRepository(ctrl)
	ctx := context.Background()

	mockRepo.EXPECT().
		List(ctx).
		Return(nil, assert.AnError)

	ucase := uc.NewListTenantsUseCase(mockRepo)

	result, err := ucase.Execute(ctx)
	assert.Error(t, err)
	assert.Nil(t, result)
}
