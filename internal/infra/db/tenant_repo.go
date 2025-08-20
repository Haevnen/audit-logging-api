package db

import (
	"context"
	"time"

	"github.com/Haevnen/audit-logging-api/internal/domain/tenant"
	"gorm.io/gorm"
)

type TenantRepo struct {
	db *gorm.DB
}

func NewTenantRepo(db *gorm.DB) *TenantRepo {
	return &TenantRepo{
		db: db,
	}
}

type tenantModel struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (tenantModel) TableName() string {
	return "tenants"
}

func (r *TenantRepo) Create(ctx context.Context, t tenant.Tenant) (tenant.Tenant, error) {
	m := tenantModel{
		ID:   t.ID,
		Name: t.Name,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return tenant.Tenant{}, err
	}
	return tenant.Tenant{
		ID:        m.ID,
		Name:      m.Name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}, nil
}

func (r *TenantRepo) List(ctx context.Context) ([]tenant.Tenant, error) {
	var models []tenantModel
	if err := r.db.WithContext(ctx).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	tenants := make([]tenant.Tenant, 0, len(models))
	for _, m := range models {
		tenants = append(tenants, tenant.Tenant{
			ID:        m.ID,
			Name:      m.Name,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}
	return tenants, nil
}
