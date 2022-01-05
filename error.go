package sirkulator

import "fmt"

// TODO read this:
// https://blog.carlmjohnson.net/post/2020/working-with-errors-as/

// Application error codes, which map nicely to http status codes
const (
	CodeInternal     = "internal"     // 500 catch-all "our" errors
	CodeConflict     = "conflict"     // 409
	CodeInvalid      = "invalid"      // 400 user errors
	CodeNotFound     = "not_found"    // 404
	CodeUnauthorized = "unauthorized" // 401

	// TODO theese also?
	// ErrTimeout	      = "timeout"            504 http.StatusGatewayTimeout
	// ErrTempUnavailable = "temp_unavaialable"  503 http.StatusServiceUnavailable
)

var (
	ErrInternal     = Error{Code: CodeInternal}
	ErrConflict     = Error{Code: CodeConflict}
	ErrInvalid      = Error{Code: CodeInvalid}
	ErrNotFound     = Error{Code: CodeNotFound}
	ErrUnauthorized = Error{Code: CodeUnauthorized}
)

type Error struct {
	Code    string
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("sirkulator: code=%s message=%s", e.Code, e.Message)
}

// Errorf returns an Error with the given code and formatted message.
func Errorf(code string, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
