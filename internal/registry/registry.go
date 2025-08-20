package registry

import (
	"github.com/Haevnen/audit-logging-api/internal/infra/db"
	"github.com/Haevnen/audit-logging-api/internal/usecase/tenant"
	"gorm.io/gorm"
)

type Registry struct {
	db *gorm.DB
}

func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{
		db: db,
	}
}

func (r *Registry) CreateTenantUseCase() *tenant.CreateTenantUseCase {
	return &tenant.CreateTenantUseCase{
		Repo: db.NewTenantRepo(r.db),
	}
}

func (r *Registry) ListTenantsUseCase() *tenant.ListTenantsUseCase {
	return &tenant.ListTenantsUseCase{
		Repo: db.NewTenantRepo(r.db),
	}
}
