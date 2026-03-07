package zinc

import (
	"errors"
	"net/http"
)

// ErrorHandler handles an error returned from a Zinc handler.
type ErrorHandler func(*Context, error)

// HTTPError represents an HTTP error with a status code and optional message.
type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return http.StatusText(e.Code)
}

func NewError(code int) *HTTPError {
	return &HTTPError{Code: code}
}

func (e *HTTPError) WithMessage(message string) *HTTPError {
	e.Message = message
	return e
}

func defaultErrorHandler(c *Context, err error) {
	if c == nil || err == nil || c.written {
		return
	}

	if httpErr, ok := err.(*HTTPError); ok {
		_ = c.Status(httpErr.Code).String(httpErr.Error())
		return
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		_ = c.Status(httpErr.Code).String(httpErr.Error())
		return
	}

	_ = c.Status(http.StatusInternalServerError).String(http.StatusText(http.StatusInternalServerError))
}

var (
	ErrBadRequest                    = NewError(StatusBadRequest)
	ErrUnauthorized                  = NewError(StatusUnauthorized)
	ErrPaymentRequired               = NewError(StatusPaymentRequired)
	ErrForbidden                     = NewError(StatusForbidden)
	ErrNotFound                      = NewError(StatusNotFound)
	ErrMethodNotAllowed              = NewError(StatusMethodNotAllowed)
	ErrNotAcceptable                 = NewError(StatusNotAcceptable)
	ErrProxyAuthRequired             = NewError(StatusProxyAuthRequired)
	ErrRequestTimeout                = NewError(StatusRequestTimeout)
	ErrConflict                      = NewError(StatusConflict)
	ErrGone                          = NewError(StatusGone)
	ErrLengthRequired                = NewError(StatusLengthRequired)
	ErrPreconditionFailed            = NewError(StatusPreconditionFailed)
	ErrRequestEntityTooLarge         = NewError(StatusRequestEntityTooLarge)
	ErrRequestURITooLong             = NewError(StatusRequestURITooLong)
	ErrUnsupportedMediaType          = NewError(StatusUnsupportedMediaType)
	ErrRequestedRangeNotSatisfiable  = NewError(StatusRequestedRangeNotSatisfiable)
	ErrExpectationFailed             = NewError(StatusExpectationFailed)
	ErrTeapot                        = NewError(StatusTeapot)
	ErrMisdirectedRequest            = NewError(StatusMisdirectedRequest)
	ErrUnprocessableEntity           = NewError(StatusUnprocessableEntity)
	ErrLocked                        = NewError(StatusLocked)
	ErrFailedDependency              = NewError(StatusFailedDependency)
	ErrTooEarly                      = NewError(StatusTooEarly)
	ErrUpgradeRequired               = NewError(StatusUpgradeRequired)
	ErrPreconditionRequired          = NewError(StatusPreconditionRequired)
	ErrTooManyRequests               = NewError(StatusTooManyRequests)
	ErrRequestHeaderFieldsTooLarge   = NewError(StatusRequestHeaderFieldsTooLarge)
	ErrUnavailableForLegalReasons    = NewError(StatusUnavailableForLegalReasons)
	ErrInternalServerError           = NewError(StatusInternalServerError)
	ErrNotImplemented                = NewError(StatusNotImplemented)
	ErrBadGateway                    = NewError(StatusBadGateway)
	ErrServiceUnavailable            = NewError(StatusServiceUnavailable)
	ErrGatewayTimeout                = NewError(StatusGatewayTimeout)
	ErrHTTPVersionNotSupported       = NewError(StatusHTTPVersionNotSupported)
	ErrVariantAlsoNegotiates         = NewError(StatusVariantAlsoNegotiates)
	ErrInsufficientStorage           = NewError(StatusInsufficientStorage)
	ErrLoopDetected                  = NewError(StatusLoopDetected)
	ErrNotExtended                   = NewError(StatusNotExtended)
	ErrNetworkAuthenticationRequired = NewError(StatusNetworkAuthenticationRequired)
)
