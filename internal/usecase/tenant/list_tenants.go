package tenant

import (
	"context"

	"github.com/Haevnen/audit-logging-api/internal/domain/tenant"
)

type ListTenantsUseCase struct {
	Repo tenant.Repository
}

func (uc *ListTenantsUseCase) Execute(ctx context.Context) ([]tenant.Tenant, error) {
	return uc.Repo.List(ctx)
}
