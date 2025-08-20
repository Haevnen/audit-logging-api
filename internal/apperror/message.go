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
		ErrInvalidKeySize:                  {httpStatus: http.StatusInternalServerError, resType: string(api.InternalError), errCode: errCodeInternalServerError, msg: "The key size is invalid."},
		ErrNotProvidedAuthenticationHeader: {httpStatus: http.StatusUnauthorized, resType: string(api.ValidationFailed), errCode: errCodeUnauthorized, msg: "The authentication header is not provided. Please provide a token."},
		ErrUnsupportedAuthorizationType:    {httpStatus: http.StatusUnauthorized, resType: string(api.ValidationFailed), errCode: errCodeUnauthorized, msg: "The authorization type is not supported. Only Bearer is supported."},
		ErrForbidden:                       {httpStatus: http.StatusForbidden, resType: string(api.PermissionDenied), errCode: errCodeForbidden, msg: "The user is forbidden to access the resource."},
		ErrInvalidRequestInput:             {httpStatus: http.StatusBadRequest, resType: string(api.ValidationFailed), errCode: errCodeInvalidRequest, msg: "The input is invalid."},
		ErrRecordNotFound:                  {httpStatus: http.StatusNotFound, resType: string(api.RequestNotFound), errCode: errCodeNotFound, msg: "The record is not found."},
		ErrInvalidPassword:                 {httpStatus: http.StatusUnauthorized, resType: string(api.ValidationFailed), errCode: errCodeUnauthorized, msg: "The password is invalid."},
		ErrExpiredRefreshToken:             {httpStatus: http.StatusUnauthorized, resType: string(api.ValidationFailed), errCode: errCodeUnauthorized, msg: "The refresh token has expired. Please login again. "},
		ErrInvalidRefreshToken:             {httpStatus: http.StatusUnauthorized, resType: string(api.ValidationFailed), errCode: errCodeUnauthorized, msg: "The refresh token is invalid. Please login again. "},
		ErrUserHasNicknameExists:           {httpStatus: http.StatusConflict, resType: string(api.ValidationFailed), errCode: errCode2000, msg: "The user has the nickname already exists."},
		ErrUserHasEmailExists:              {httpStatus: http.StatusConflict, resType: string(api.ValidationFailed), errCode: errCode2001, msg: "The user has the email already exists."},
		ErrClientHasIDExists:               {httpStatus: http.StatusConflict, resType: string(api.ValidationFailed), errCode: errCode3000, msg: "The client has the id already exists."},
		ErrUserNotExists:                   {httpStatus: http.StatusBadRequest, resType: string(api.ValidationFailed), errCode: errCode2002, msg: "The user didn't exists."},
		ErrTicketNotFound:                  {httpStatus: http.StatusNotFound, resType: string(api.RequestNotFound), errCode: errCode2003, msg: "The ticket is not found. "},
		ErrTicketHasBeenDeleted:            {httpStatus: http.StatusNotFound, resType: string(api.RequestNotFound), errCode: errCode2008, msg: "The ticket has been deleted. "},
		ErrQCNameNotExists:                 {httpStatus: http.StatusNotFound, resType: string(api.RequestNotFound), errCode: errCode2004, msg: "The QC name is not found. "},
		ErrEditorNameNotExists:             {httpStatus: http.StatusNotFound, resType: string(api.RequestNotFound), errCode: errCode2005, msg: "The editor name is not found. "},
		ErrPermissionDenied:                {httpStatus: http.StatusBadRequest, resType: string(api.ValidationFailed), errCode: errCode2006, msg: "The user is not allowed to update or remove resource "},
		ErrViewPermissionDenied:            {httpStatus: http.StatusBadRequest, resType: string(api.ValidationFailed), errCode: errCode2007, msg: "The user is not allowed to view resource "},
		ErrExportTimeOverRange:             {httpStatus: http.StatusBadRequest, resType: string(api.ValidationFailed), errCode: errCode4000, msg: "The range over 60 days."},
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
	errCode2000     = "ERR_2000"
	errCode2001     = "ERR_2001"
	errCode2002     = "ERR_2002"
	errCode2003     = "ERR_2003"
	errCode2004     = "ERR_2004"
	errCode2005     = "ERR_2005"
	errCode2006     = "ERR_2006"
	errCode2007     = "ERR_2007"
	errCode2008     = "ERR_2008"

	errCode3000 = "ERR_3000"

	errCode4000 = "ERR_4000"
)
