package tenant

import "time"

type Tenant struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
