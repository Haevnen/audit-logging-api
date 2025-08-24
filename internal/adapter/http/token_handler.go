package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/auth"
	"github.com/Haevnen/audit-logging-api/internal/registry"
)

type TokenHandler struct {
	JwtManager auth.ManagerInterface
}

func newTokenHandler(registry *registry.Registry) TokenHandler {
	return TokenHandler{
		JwtManager: registry.Manager(),
	}
}

// GenerateToken (POST /auth/token)
func (t TokenHandler) GenerateToken(c *gin.Context) {
	var genTokenReqBody api_service.GenerateTokenRequestBody
	if BindRequestBody(c, &genTokenReqBody) != nil {
		SendError(c, "invalid request body", apperror.ErrInvalidRequestInput)
		return
	}

	role := auth.Role(genTokenReqBody.Role)
	if !role.IsValid() {
		SendError(c, "invalid role", apperror.ErrInvalidRequestInput)
		return
	}

	token, err := t.JwtManager.GenerateToken(genTokenReqBody.UserId, genTokenReqBody.TenantId, role, 1*time.Hour)
	if err != nil {
		SendError(c, "failed to generate token", apperror.ErrInternalServer)
		return
	}

	c.JSON(http.StatusOK, api_service.GenerateTokenResponse{
		Token: token,
	})
}
