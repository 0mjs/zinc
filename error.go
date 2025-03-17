package zinc

import (
	"net/http"
)

// HTTPError represents an HTTP error with a status code and optional message.
type HTTPError struct {
	Code    int    // HTTP status code
	Message string // Error message
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return http.StatusText(e.Code)
}

// NewError creates a new HTTPError with the given status code.
func NewError(code int) *HTTPError {
	return &HTTPError{
		Code: code,
	}
}

// WithMessage sets a custom message for the error.
func (e *HTTPError) WithMessage(message string) *HTTPError {
	e.Message = message
	return e
}

var (
	ErrBadRequest                    = NewError(StatusBadRequest)                    // RFC 7231, 6.5.1
	ErrUnauthorized                  = NewError(StatusUnauthorized)                  // RFC 7235, 3.1
	ErrPaymentRequired               = NewError(StatusPaymentRequired)               // RFC 7231, 6.5.2
	ErrForbidden                     = NewError(StatusForbidden)                     // RFC 7231, 6.5.3
	ErrNotFound                      = NewError(StatusNotFound)                      // RFC 7231, 6.5.4
	ErrMethodNotAllowed              = NewError(StatusMethodNotAllowed)              // RFC 7231, 6.5.5
	ErrNotAcceptable                 = NewError(StatusNotAcceptable)                 // RFC 7231, 6.5.6
	ErrProxyAuthRequired             = NewError(StatusProxyAuthRequired)             // RFC 7235, 3.2
	ErrRequestTimeout                = NewError(StatusRequestTimeout)                // RFC 7231, 6.5.7
	ErrConflict                      = NewError(StatusConflict)                      // RFC 7231, 6.5.8
	ErrGone                          = NewError(StatusGone)                          // RFC 7231, 6.5.9
	ErrLengthRequired                = NewError(StatusLengthRequired)                // RFC 7231, 6.5.10
	ErrPreconditionFailed            = NewError(StatusPreconditionFailed)            // RFC 7232, 4.2
	ErrRequestEntityTooLarge         = NewError(StatusRequestEntityTooLarge)         // RFC 7231, 6.5.11
	ErrRequestURITooLong             = NewError(StatusRequestURITooLong)             // RFC 7231, 6.5.12
	ErrUnsupportedMediaType          = NewError(StatusUnsupportedMediaType)          // RFC 7231, 6.5.13
	ErrRequestedRangeNotSatisfiable  = NewError(StatusRequestedRangeNotSatisfiable)  // RFC 7233, 4.4
	ErrExpectationFailed             = NewError(StatusExpectationFailed)             // RFC 7231, 6.5.14
	ErrTeapot                        = NewError(StatusTeapot)                        // RFC 7168, 2.3.3
	ErrMisdirectedRequest            = NewError(StatusMisdirectedRequest)            // RFC 7540, 9.1.2
	ErrUnprocessableEntity           = NewError(StatusUnprocessableEntity)           // RFC 4918, 11.2
	ErrLocked                        = NewError(StatusLocked)                        // RFC 4918, 11.3
	ErrFailedDependency              = NewError(StatusFailedDependency)              // RFC 4918, 11.4
	ErrTooEarly                      = NewError(StatusTooEarly)                      // RFC 8470, 5.2.
	ErrUpgradeRequired               = NewError(StatusUpgradeRequired)               // RFC 7231, 6.5.15
	ErrPreconditionRequired          = NewError(StatusPreconditionRequired)          // RFC 6585, 3
	ErrTooManyRequests               = NewError(StatusTooManyRequests)               // RFC 6585, 4
	ErrRequestHeaderFieldsTooLarge   = NewError(StatusRequestHeaderFieldsTooLarge)   // RFC 6585, 5
	ErrUnavailableForLegalReasons    = NewError(StatusUnavailableForLegalReasons)    // RFC 7725, 3
	ErrInternalServerError           = NewError(StatusInternalServerError)           // RFC 7231, 6.6.1
	ErrNotImplemented                = NewError(StatusNotImplemented)                // RFC 7231, 6.6.2
	ErrBadGateway                    = NewError(StatusBadGateway)                    // RFC 7231, 6.6.3
	ErrServiceUnavailable            = NewError(StatusServiceUnavailable)            // RFC 7231, 6.6.4
	ErrGatewayTimeout                = NewError(StatusGatewayTimeout)                // RFC 7231, 6.6.5
	ErrHTTPVersionNotSupported       = NewError(StatusHTTPVersionNotSupported)       // RFC 7231, 6.6.6
	ErrVariantAlsoNegotiates         = NewError(StatusVariantAlsoNegotiates)         // RFC 2295, 8.1
	ErrInsufficientStorage           = NewError(StatusInsufficientStorage)           // RFC 4918, 11.5
	ErrLoopDetected                  = NewError(StatusLoopDetected)                  // RFC 5842, 7.2
	ErrNotExtended                   = NewError(StatusNotExtended)                   // RFC 2774, 7
	ErrNetworkAuthenticationRequired = NewError(StatusNetworkAuthenticationRequired) // RFC 6585, 6
)
