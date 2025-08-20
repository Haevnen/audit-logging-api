package registry

import (
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/infra/db"
	"github.com/Haevnen/audit-logging-api/internal/usecase/tenant"
	"gorm.io/gorm"
)

type Registry struct {
	db  *gorm.DB
	key string
}

func NewRegistry(db *gorm.DB, key string) *Registry {
	return &Registry{
		db:  db,
		key: key,
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

func (r *Registry) Manager() *auth.Manager {
	return auth.NewManager(r.key)
}
