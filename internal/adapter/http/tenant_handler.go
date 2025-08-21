package handler

import (
	"net/http"

	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/Haevnen/audit-logging-api/internal/usecase/tenant"
	"github.com/gin-gonic/gin"
)

type tenantHandler struct {
	CreateUC *tenant.CreateTenantUseCase
	ListUC   *tenant.ListTenantsUseCase
}

func newTenantHandler(r *registry.Registry) tenantHandler {
	return tenantHandler{
		CreateUC: r.CreateTenantUseCase(),
		ListUC:   r.ListTenantsUseCase(),
	}
}

// (GET /tenants)
func (h tenantHandler) ListTenants(g *gin.Context) {
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
func (h tenantHandler) CreateTenant(g *gin.Context) {
	var body api_service.CreateTenantRequestBody
	if err := bindRequestBody(g, &body); err != nil {
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
