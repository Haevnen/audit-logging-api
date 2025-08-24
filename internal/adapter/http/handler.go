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
	TenantHandler
	TokenHandler
	LogHandler
	LogStreamHandler
}

func New(r *registry.Registry) Handler {
	h := Handler{}
	h.TenantHandler = newTenantHandler(r)
	h.TokenHandler = newTokenHandler(r)
	h.LogHandler = newLogHandler(r)
	h.LogStreamHandler = newLogStreamHandler(r)
	return h
}

// GetPing (GET /ping)
func (Handler) GetPing(ctx *gin.Context) {
	resp := api_service.Pong{
		Ping: "pong",
	}

	ctx.JSON(http.StatusOK, resp)
}

func BindRequestBody(ctx *gin.Context, body interface{}) error {
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
