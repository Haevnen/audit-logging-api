package handler

import (
	"net/http"

	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/apperror"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/gin-gonic/gin"
)

const (
	DateTimeFormat = "2006-01-02T15:04:05.000Z"
)

type Handler struct {
	tenantHandler
	tokenHandler
	logHandler
}

func New(r *registry.Registry) Handler {
	h := Handler{}
	h.tenantHandler = newTenantHandler(r)
	h.tokenHandler = newTokenHandler(r)
	h.logHandler = newLogHandler(r)
	return h
}

// GetPing (GET /ping)
func (Handler) GetPing(ctx *gin.Context) {
	resp := api_service.Pong{
		Ping: "pong",
	}

	ctx.JSON(http.StatusOK, resp)
}

func bindRequestBody(ctx *gin.Context, body interface{}) error {
	if err := ctx.ShouldBindJSON(body); err != nil {
		return err
	}
	return nil
}

func SendError(ctx *gin.Context, title string, err error) {
	appErr := apperror.New(ctx, err)
	ctx.JSON(appErr.HTTPStatus(), api_service.Error{
		Type:   api_service.ErrorType(appErr.ResType()),
		Title:  title,
		Code:   appErr.ErrorCode(),
		Detail: appErr.Detail(),
	})
}
