package tenant

//go:generate mockgen -source=interface.go -destination=./mocks/mock_usecase.go -package=mocks
import (
	"context"

	entitytenant "github.com/Haevnen/audit-logging-api/internal/entity/tenant"
)

// CreateTenantUseCaseInterface defines behavior for creating tenants.
type CreateTenantUseCaseInterface interface {
	Execute(ctx context.Context, name string) (*entitytenant.Tenant, error)
}

// ListTenantsUseCaseInterface defines behavior for listing tenants.
type ListTenantsUseCaseInterface interface {
	Execute(ctx context.Context) ([]entitytenant.Tenant, error)
}
