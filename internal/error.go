// MIT License

// Copyright (c) 2020 Ben Johnson

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Source: https://github.com/benbjohnson/wtf/blob/main/error.go
package echosight

import (
	"errors"
	"fmt"
)

// Application error codes.
const (
	ECONFLICT       string = "conflict"
	EINTERNAL       string = "internal"
	EINVALID        string = "invalid"
	ENOTFOUND       string = "not_found"
	ENOTIMPLEMENTED string = "not_implemented"
	EUNAUTHORIZED   string = "unauthorized"
)

// Error represents an application-specific error. Application errors can be
// unwrapped by the caller to extract out the code & message.
//
// Any non-application error (such as a disk error) should be reported as an
// EINTERNAL error and the human user should only see "Internal error" as the
// message. These low-level internal error details should only be logged and
// reported to the operator of the application (not the end user).
type Error struct {
	// Machine-readable error code.
	Code string
	// Human-readable error message.
	Message string

	// Internal error message - schould not exposed to API user
	Internal error
	// Data can be use in http response body
	Data any
}

// Error implements the error interface. Not used by the application otherwise.
func (e *Error) Error() string {
	return fmt.Sprintf("%s code=%s", e.String(), e.Code)
}

// String returns the whole error message inclusive internal error if exists
func (e *Error) String() string {
	if e.Internal != nil && e.Message != e.Internal.Error() {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}

	return e.Message
}

// WithError wraps the error to internal
func (e *Error) WithError(err error) *Error {
	e.Internal = err
	return e
}

func (e *Error) WithData(data any) *Error {
	e.Data = data
	return e
}

// ErrorCode unwraps an application error and returns its code.
// Non-application errors always return EINTERNAL.
func ErrorCode(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Code
	}
	return EINTERNAL
}

// ErrorMessage unwraps an application error and returns its message.
// Non-application errors always return "Internal error".
func ErrorMessage(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Message
	}
	return "Internal error."
}

// FromError creates an Application Error from an error interface
func FromError(err error) *Error {
	var e *Error
	if err != nil && errors.As(err, &e) {
		return e
	} else if err != nil {
		return &Error{
			Code:    EINTERNAL,
			Message: err.Error(),
			Data:    nil,
		}
	}

	return &Error{
		Code:    EINTERNAL,
		Message: "Internal error",
		Data:    nil,
	}
}

// Errorf is a helper function to return an Error with a given code and formatted message.
func Errorf(code string, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func ErrInternalf(format string, args ...any) *Error {
	return Errorf(EINTERNAL, format, args...)
}

// ErrInvalidf
// HTTP: BadRequest
func ErrInvalidf(format string, args ...any) *Error {
	return Errorf(EINVALID, format, args...)
}

func ErrConflictf(format string, args ...any) *Error {
	return Errorf(ECONFLICT, format, args...)
}

func ErrNotfoundf(format string, args ...any) *Error {
	return Errorf(ENOTFOUND, format, args...)
}

func ErrUnauthorizedf(format string, args ...any) *Error {
	return Errorf(EUNAUTHORIZED, format, args...)
}

func ErrNotImplementedf(format string, args ...any) *Error {
	return Errorf(ENOTIMPLEMENTED, format, args...)
}
