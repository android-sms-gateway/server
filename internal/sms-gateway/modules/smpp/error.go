package smpp

import (
	"errors"
	"net/http"
	"strings"
)

// SMPP Error Codes - Standard SMPP v3.4 error codes
// These are already defined in session.go, but we re-export them here for clarity
const (
	ErrNoError           uint32 = 0x00000000 // ESME_ROK
	ErrInvalidMsgLen     uint32 = 0x00000001 // ESME_RINVMSGLEN
	ErrInvalidCmdLen     uint32 = 0x00000002 // ESME_RINVCMDLEN
	ErrInvalidCmdID      uint32 = 0x00000003 // ESME_RINVCMDID
	ErrInvalidBindStatus uint32 = 0x0000000D // ESME_RINVBNDSTS
	ErrAccessFail        uint32 = 0x00000009 // ESME_RACCESSFAIL
	ErrNotFound          uint32 = 0x0000000A // ESME_RNOTFOUND
	ErrAlreadyBound      uint32 = 0x0000000F // ESME_RALREADYBINDSYS
	ErrSystemErr         uint32 = 0x00000014 // ESME_RSYSERR
	ErrInvalidParam      uint32 = 0x0000001B // ESME_RINVPARAM
)

// ErrorMapping maps an HTTP status code to an SMPP error code
type ErrorMapping struct {
	HTTPStatus  int
	SMPPErrCode uint32
	Description string
}

// ErrorMappings is the canonical mapping table from HTTP errors to SMPP error codes
var ErrorMappings = []ErrorMapping{
	{HTTPStatus: http.StatusUnauthorized, SMPPErrCode: ErrAccessFail, Description: "Authentication failed"},
	{HTTPStatus: http.StatusBadRequest, SMPPErrCode: ErrInvalidParam, Description: "Invalid parameter"},
	{HTTPStatus: http.StatusNotFound, SMPPErrCode: ErrNotFound, Description: "Resource not found"},
	{HTTPStatus: http.StatusConflict, SMPPErrCode: ErrAlreadyBound, Description: "Already bound"},
	{HTTPStatus: http.StatusInternalServerError, SMPPErrCode: ErrSystemErr, Description: "System error"},
	{HTTPStatus: http.StatusServiceUnavailable, SMPPErrCode: ErrSystemErr, Description: "Service unavailable"},
	{HTTPStatus: http.StatusGatewayTimeout, SMPPErrCode: ErrSystemErr, Description: "Gateway timeout"},
	{HTTPStatus: http.StatusBadGateway, SMPPErrCode: ErrSystemErr, Description: "Bad gateway"},
	{HTTPStatus: http.StatusForbidden, SMPPErrCode: ErrAccessFail, Description: "Access forbidden"},
	{HTTPStatus: http.StatusTooManyRequests, SMPPErrCode: ErrSystemErr, Description: "Rate limited"},
	{HTTPStatus: http.StatusRequestTimeout, SMPPErrCode: ErrSystemErr, Description: "Request timeout"},
	{HTTPStatus: http.StatusNotImplemented, SMPPErrCode: ErrInvalidCmdID, Description: "Not implemented"},
}

// httpToSMPPErrMap is a pre-computed map for O(1) lookup
var httpToSMPPErrMap map[int]uint32

func init() {
	httpToSMPPErrMap = make(map[int]uint32, len(ErrorMappings))
	for _, m := range ErrorMappings {
		httpToSMPPErrMap[m.HTTPStatus] = m.SMPPErrCode
	}
}

// MapHTTPErrorToSMPP converts an HTTP status code to the corresponding SMPP error code.
// Returns ESME_ROK (0) if the status indicates success (< 400).
// Returns ESME_RSYSERR for unmapped error codes.
func MapHTTPErrorToSMPP(httpStatus int) uint32 {
	if httpStatus < 400 {
		return ErrNoError
	}
	if code, ok := httpToSMPPErrMap[httpStatus]; ok {
		return code
	}
	// Default for unmapped 4xx/5xx errors
	if httpStatus >= 400 && httpStatus < 500 {
		return ErrInvalidParam
	}
	return ErrSystemErr
}

// MapErrorToSMPP attempts to extract an HTTP status code from an error and map it to SMPP.
// It checks for errors that implement StatusCode() (like *Error type) or wraps an HTTP error.
// Falls back to ErrSystemErr if no mapping is found.
func MapErrorToSMPP(err error) uint32 {
	if err == nil {
		return ErrNoError
	}

	// Check if it's an HTTPError with a status code
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return MapHTTPErrorToSMPP(httpErr.StatusCode)
	}

	// Check for wrapped errors with status code
	type statusCode interface {
		StatusCode() int
	}
	var sc statusCode
	if errors.As(err, &sc) {
		return MapHTTPErrorToSMPP(sc.StatusCode())
	}

	// Try to extract status from error message (fallback)
	// e.g., "401 Unauthorized" or "HTTP 404"
	errMsg := err.Error()
	for _, m := range ErrorMappings {
		if strings.Contains(errMsg, http.StatusText(m.HTTPStatus)) ||
			strings.Contains(errMsg, string(rune(m.HTTPStatus))) {
			return m.SMPPErrCode
		}
	}

	// Default to system error
	return ErrSystemErr
}

// HTTPError represents an error with an HTTP status code
type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTPError with the given status code and message
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewHTTPErrorf creates a new HTTPError wrapping another error
func NewHTTPErrorf(statusCode int, format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    format,
	}
}

// SMPPErrCode returns the SMPP error code for this HTTP error
func (e *HTTPError) SMPPErrCode() uint32 {
	return MapHTTPErrorToSMPP(e.StatusCode)
}

// SMPPError returns the SMPP error description for a given error code
func SMPPError(code uint32) string {
	switch code {
	case ErrNoError:
		return "OK"
	case ErrInvalidMsgLen:
		return "Invalid Message Length"
	case ErrInvalidCmdLen:
		return "Invalid Command Length"
	case ErrInvalidCmdID:
		return "Invalid Command ID"
	case ErrInvalidBindStatus:
		return "Invalid Binding Status"
	case ErrAccessFail:
		return "Access Failed"
	case ErrNotFound:
		return "Not Found"
	case ErrAlreadyBound:
		return "Already Bound"
	case ErrSystemErr:
		return "System Error"
	case ErrInvalidParam:
		return "Invalid Parameter"
	default:
		return "Unknown Error"
	}
}
