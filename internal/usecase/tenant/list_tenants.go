package tenant

import (
	"context"

	"github.com/Haevnen/audit-logging-api/internal/entity/tenant"
	"github.com/Haevnen/audit-logging-api/internal/repository"
)

type ListTenantsUseCase struct {
	Repo repository.TenantRepository
}

func NewListTenantsUseCase(repo repository.TenantRepository) *ListTenantsUseCase {
	return &ListTenantsUseCase{Repo: repo}
}

func (uc *ListTenantsUseCase) Execute(ctx context.Context) ([]tenant.Tenant, error) {
	return uc.Repo.List(ctx)
}
