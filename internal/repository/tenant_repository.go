package repository

import (
	"context"

	"gorm.io/gorm"

	entity "github.com/Haevnen/audit-logging-api/internal/entity/tenant"
)

type TenantRepository interface {
	Create(ctx context.Context, t *entity.Tenant) (*entity.Tenant, error)
	List(ctx context.Context) ([]entity.Tenant, error)
}

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *tenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	if err := r.db.WithContext(ctx).Create(tenant).Error; err != nil {
		return nil, err
	}
	return tenant, nil
}

func (r *tenantRepository) List(ctx context.Context) ([]entity.Tenant, error) {
	var tenants []entity.Tenant
	err := r.db.WithContext(ctx).Order("created_at asc").Find(&tenants).Error
	return tenants, err
}
