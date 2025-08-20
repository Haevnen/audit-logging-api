package tenant

import "context"

type Repository interface {
	Create(ctx context.Context, t Tenant) (Tenant, error)
	List(ctx context.Context) ([]Tenant, error)
}
