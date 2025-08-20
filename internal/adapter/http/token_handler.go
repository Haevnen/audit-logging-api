package handler

import (
	"net/http"
	"time"

	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/gin-gonic/gin"
)

type tokenHandler struct {
	jwtManager *auth.Manager
}

func newTokenHandler(registry *registry.Registry) tokenHandler {
	return tokenHandler{
		jwtManager: registry.Manager(),
	}
}

// GenerateToken (POST /auth/token)
func (t tokenHandler) GenerateToken(c *gin.Context) {
	var genTokenReqBody api_service.GenerateTokenRequestBody
	if bindRequestBody(c, &genTokenReqBody) != nil {
		SendError(c, "invalid request body", apperror.ErrInvalidRequestInput)
		return
	}

	role := auth.Role(genTokenReqBody.Role)
	if !role.IsValid() {
		SendError(c, "invalid role", apperror.ErrInvalidRequestInput)
		return
	}

	var (
		userId   string
		tenantId string
	)

	if genTokenReqBody.TenantId != nil {
		tenantId = *genTokenReqBody.TenantId
	}

	if genTokenReqBody.UserId != nil {
		userId = *genTokenReqBody.UserId
	}

	token, err := t.jwtManager.GenerateToken(userId, tenantId, role, 1*time.Hour)
	if err != nil {
		SendError(c, "failed to generate token", apperror.ErrInternalServer)
		return
	}

	c.JSON(http.StatusOK, api_service.GenerateTokenResponse{
		Token: token,
	})
}
