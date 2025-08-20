package apperror

import (
	"context"
	"fmt"
	"net/http"

	api "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
)

type definedErrorDetail struct {
	httpStatus int
	resType    string
	errCode    string
	msg        string
}

var (
	errMessageMap = map[error]definedErrorDetail{

		context.Canceled:                   {httpStatus: http.StatusBadRequest, resType: string(api.ValidationFailed), errCode: errCode4999, msg: "The operation was canceled."},
		ErrInternalServer:                  {httpStatus: http.StatusInternalServerError, resType: string(api.InternalError), errCode: errCodeInternalServerError, msg: "An Unexpected Error has occurred."},
		ErrInvalidToken:                    {httpStatus: http.StatusUnauthorized, resType: string(api.ValidationFailed), errCode: errCodeUnauthorized, msg: "The token is invalid."},
		ErrTokenExpired:                    {httpStatus: http.StatusUnauthorized, resType: string(api.ValidationFailed), errCode: errCodeUnauthorized, msg: "The token has expired."},
		ErrNotProvidedAuthenticationHeader: {httpStatus: http.StatusUnauthorized, resType: string(api.ValidationFailed), errCode: errCodeUnauthorized, msg: "The authentication header is not provided. Please provide a token."},
		ErrUnsupportedAuthorizationType:    {httpStatus: http.StatusUnauthorized, resType: string(api.ValidationFailed), errCode: errCodeUnauthorized, msg: "The authorization type is not supported. Only Bearer is supported."},
		ErrForbidden:                       {httpStatus: http.StatusForbidden, resType: string(api.PermissionDenied), errCode: errCodeForbidden, msg: "The user is forbidden to access the resource."},
		ErrInvalidRequestInput:             {httpStatus: http.StatusBadRequest, resType: string(api.ValidationFailed), errCode: errCodeInvalidRequest, msg: "The input is invalid."},
		ErrRecordNotFound:                  {httpStatus: http.StatusNotFound, resType: string(api.RequestNotFound), errCode: errCodeNotFound, msg: "The record is not found."},
	}
)

func (d definedErrorDetail) detailsEN(params ...any) []string {
	return []string{fmt.Sprintf(d.msg, params...)}
}

const (
	errCode4999                = "ERR_4999"
	errCodeInternalServerError = "ERR_9999"
	errCodeUnauthorized        = "ERR_401"
	errCodeForbidden           = "ERR_403"

	errCodeInvalidRequest = "ERR_400"

	errCodeNotFound = "ERR_404"
)
