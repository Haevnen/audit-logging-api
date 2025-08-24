package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/Haevnen/audit-logging-api/internal/usecase/tenant"
)

type TenantHandler struct {
	CreateUC tenant.CreateTenantUseCaseInterface
	ListUC   tenant.ListTenantsUseCaseInterface
}

func newTenantHandler(r *registry.Registry) TenantHandler {
	return TenantHandler{
		CreateUC: r.CreateTenantUseCase(),
		ListUC:   r.ListTenantsUseCase(),
	}
}

// (GET /tenants)
func (h TenantHandler) ListTenants(g *gin.Context) {
	tenants, err := h.ListUC.Execute(g.Request.Context())
	if err != nil {
		SendError(g, err.Error(), apperror.ErrInternalServer)
		return
	}

	resp := make([]api_service.Tenant, 0, len(tenants))
	for _, t := range tenants {
		resp = append(resp, api_service.Tenant{
			Id:        t.ID,
			Name:      t.Name,
			CreatedAt: t.CreatedAt.Format(DateTimeFormat),
			UpdatedAt: t.UpdatedAt.Format(DateTimeFormat),
		})
	}
	g.JSON(http.StatusOK, resp)
}

// (POST /tenants)
func (h TenantHandler) CreateTenant(g *gin.Context) {
	var body api_service.CreateTenantRequestBody
	if err := BindRequestBody(g, &body); err != nil {
		SendError(g, err.Error(), apperror.ErrInvalidRequestInput)
		return
	}

	if len(body.Name) == 0 {
		SendError(g, "name is required", apperror.ErrInvalidRequestInput)
		return
	}

	t, err := h.CreateUC.Execute(g.Request.Context(), body.Name)
	if err != nil {
		SendError(g, err.Error(), apperror.ErrInternalServer)
		return
	}

	resp := api_service.Tenant{
		Id:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt.Format(DateTimeFormat),
		UpdatedAt: t.UpdatedAt.Format(DateTimeFormat),
	}
	g.JSON(http.StatusCreated, resp)
}
