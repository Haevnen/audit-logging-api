package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/Haevnen/audit-logging-api/internal/entity/tenant"
	"github.com/Haevnen/audit-logging-api/internal/repository"
)

type CreateTenantUseCase struct {
	Repo repository.TenantRepository
}

func NewCreateTenantUseCase(repo repository.TenantRepository) *CreateTenantUseCase {
	return &CreateTenantUseCase{Repo: repo}
}

func (uc *CreateTenantUseCase) Execute(ctx context.Context, name string) (*tenant.Tenant, error) {
	t := &tenant.Tenant{
		ID:   uuid.New().String(),
		Name: name,
	}
	return uc.Repo.Create(ctx, t)
}
