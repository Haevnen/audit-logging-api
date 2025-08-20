package tenant

import (
	"context"

	"github.com/Haevnen/audit-logging-api/internal/domain/tenant"
	"github.com/google/uuid"
)

type CreateTenantUseCase struct {
	Repo tenant.Repository
}

func (uc *CreateTenantUseCase) Execute(ctx context.Context, name string) (tenant.Tenant, error) {
	t := tenant.Tenant{
		ID:   uuid.NewString(),
		Name: name,
	}
	return uc.Repo.Create(ctx, t)
}
