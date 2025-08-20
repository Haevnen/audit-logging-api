// Package apperror application errors
// Basically we wrap every error in the application layer (except for util) by apperror.New
// When returning the response, it sets http status, error code, detail via apperror.Error
package apperror

import (
	"context"
	"errors"
	"net/http"
	"strings"

	apiModel "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
)

// Error error for application
type Error struct {
	err        error
	httpStatus int
	resType    string
	errCode    string
	details    []string // Use slice in case several messages are returned in the response
}

// Define application layer errors
var (
	ErrInternalServer                   = errors.New("ERR_INTERNAL_SERVER")
	ErrInvalidToken                     = errors.New("ERR_INVALID_TOKEN")
	ErrTokenExpired                     = errors.New("ERR_TOKEN_EXPIRED")
	ErrInvalidKeySize                   = errors.New("ERR_INVALID_KEY_SIZE")
	ErrNotProvidedAuthenticationHeader  = errors.New("ERR_NOT_PROVIDED_AUTHENTICATION_HEADER")
	ErrInvalidAuthorizationHeaderFormat = errors.New("ERR_INVALID_AUTHORIZATION_HEADER_FORMAT")
	ErrUnsupportedAuthorizationType     = errors.New("ERR_UNSUPPORTED_AUTHORIZATION_TYPE")
	ErrForbidden                        = errors.New("ERR_FORBIDDEN")
	ErrInvalidRequestInput              = errors.New("ERR_INVALID_REQUEST_INPUT")
	ErrRecordNotFound                   = errors.New("ERR_RECORD_NOT_FOUND")
	ErrInvalidPassword                  = errors.New("ERR_INVALID_PASSWORD")
	ErrUserHasNicknameExists            = errors.New("ERR_USER_HAS_NICKNAME_EXISTS")
	ErrUserHasEmailExists               = errors.New("ERR_USER_HAS_EMAIL_EXISTS")
	ErrExpiredRefreshToken              = errors.New("ERR_EXPIRED_REFRESH_TOKEN")
	ErrInvalidRefreshToken              = errors.New("ERR_INVALID_REFRESH_TOKEN")
	ErrClientHasIDExists                = errors.New("ERR_CLIENT_HAS_ID_EXISTS")
	ErrUserNotExists                    = errors.New("ERR_USER_NOT_EXISTS")
	ErrTicketNotFound                   = errors.New("ERR_TICKET_NOT_FOUND")
	ErrQCNameNotExists                  = errors.New("ERR_QC_NAME_NOT_EXISTS")
	ErrEditorNameNotExists              = errors.New("ERR_EDITOR_NAME_NOT_EXISTS")
	ErrPermissionDenied                 = errors.New("ERR_PERMISSION_DENIED")
	ErrViewPermissionDenied             = errors.New("ERR_VIEW_PERMISSION_DENIED")
	ErrTicketHasBeenDeleted             = errors.New("ERR_TICKET_HAS_BEEN_DELETED")
	ErrExportTimeOverRange              = errors.New("ERR_EXPORT_TIME_OVER_RANGE")
)

func New(_ context.Context, err error, params ...any) *Error {
	if err == nil {
		return nil
	}

	e := &Error{}
	if errors.As(err, &e) {
		return e
	}

	e.err = err
	if defined, ok := errMessageMap[err]; ok {
		e.errCode = err.Error()
		if defined.errCode != "" {
			e.errCode = defined.errCode
		}
		e.httpStatus = defined.httpStatus
		e.resType = defined.resType
		e.details = defined.detailsEN(params...)
	} else {
		if errors.Is(err, context.Canceled) {
			e.errCode = errMessageMap[context.Canceled].errCode
			e.httpStatus = errMessageMap[context.Canceled].httpStatus
			e.resType = errMessageMap[context.Canceled].resType
			e.details = []string{errMessageMap[context.Canceled].msg}
		} else {
			e.errCode = errCodeInternalServerError
			e.httpStatus = http.StatusInternalServerError
			e.resType = string(apiModel.InternalError)
			e.details = []string{err.Error()}
		}
	}

	return e
}

// Error satisfies error interface.
func (e *Error) Error() string {
	// Please not use this method for returning error to the client
	return e.err.Error()
}

// HTTPStatus http status code
func (e *Error) HTTPStatus() int {
	return e.httpStatus
}

// ErrorCode error code for response
func (e *Error) ErrorCode() string {
	return e.errCode
}

// ResType response type for response
func (e *Error) ResType() string {
	return e.resType
}

// Detail error message for response
func (e *Error) Detail() string {
	return strings.Join(e.details, ",")
}
