package registry

import (
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/interactor"
	"github.com/Haevnen/audit-logging-api/internal/repository"
	"github.com/Haevnen/audit-logging-api/internal/usecase/log"
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
	return tenant.NewCreateTenantUseCase(repository.NewTenantRepository(r.db))

}

func (r *Registry) ListTenantsUseCase() *tenant.ListTenantsUseCase {
	return tenant.NewListTenantsUseCase(repository.NewTenantRepository(r.db))
}

func (r *Registry) CreateLogUseCase() *log.CreateLogUseCase {
	return log.NewCreateLogUseCase(repository.NewLogRepository(r.db), r.TxManager())
}

func (r *Registry) GetLogUseCase() *log.GetLogUseCase {
	return log.NewGetLogUseCase(repository.NewLogRepository(r.db))
}

func (r *Registry) Manager() *auth.Manager {
	return auth.NewManager(r.key)
}

func (r *Registry) TxManager() *interactor.TxManager {
	return interactor.NewTxManager(r.db)
}
