package error

import (
	"encoding/json"
)

// Code ...
type Code int

// String ...
func (c Code) String() string {
	return string(c)
}

// Error ...
type Error struct {
	Err     error  `json:"-"`
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// Error returns a human readable error.
func (e Error) Error() string {
	if data, err := json.Marshal(e); err == nil {
		return string(data)
	}
	return "{}"
}

// WithMessage ...
func (e Error) WithMessage(msg string) Error {
	return Error{
		Err:     e.Err,
		Code:    e.Code,
		Message: msg,
		Detail:  e.Detail,
	}
}

// Unwrap ...
func (e Error) Unwrap() error { return e.Err }

// Errors ...
type Errors []error

var _ error = Errors{}

// MarshalJSON converts slice of error
func (errs Errors) Error() string {
	var tmpErrs struct {
		Errors []Error `json:"errors,omitempty"`
	}

	for _, daErr := range errs {
		var err error
		switch daErr.(type) {
		case Error:
			err = daErr.(Error)
		default:
			err = UnknownError(daErr).(Error).WithMessage(err.Error())
		}
		tmpErrs.Errors = append(tmpErrs.Errors, Error{
			Err:     err.(Error).Err,
			Code:    err.(Error).Code,
			Message: err.(Error).Message,
			Detail:  err.(Error).Detail,
		})
	}

	if msg, err := json.Marshal(tmpErrs); err == nil {
		return string(msg)
	}

	return "{}"
}

// Len returns the current number of errors.
func (errs Errors) Len() int {
	return len(errs)
}

// NewErrs ...
func NewErrs(err error) Errors {
	return Errors{err}
}

const (
	// ObjectNotFoundErrorCode is code for the error of no object found
	ObjectNotFoundErrorCode = 10000 + iota
	// ObjectConflictErrorCode ...
	ObjectConflictErrorCode
	// UnAuthorizedErrorCode ...
	UnAuthorizedErrorCode
	// BadRequestErrorCode ...
	BadRequestErrorCode
	// ForbiddenErrorCode ...
	ForbiddenErrorCode
	// PreconditionErrorCode ...
	PreconditionErrorCode
	// GeneralErrorCode ...
	GeneralErrorCode
)

// NotFoundError is error for the case of object not found
func NotFoundError(err error) error {
	return Error{
		Err:     err,
		Code:    ObjectNotFoundErrorCode,
		Message: err.Error(),
		Detail:  "not found",
	}
}

// ConflictError is error for the case of object conflict
func ConflictError(err error) error {
	return Error{
		Err:     err,
		Code:    ObjectConflictErrorCode,
		Message: err.Error(),
		Detail:  "conflict",
	}
}

// UnauthorizedError is error for the case of unauthorized accessing
func UnauthorizedError(err error) error {
	return Error{
		Err:     err,
		Code:    UnAuthorizedErrorCode,
		Message: err.Error(),
		Detail:  "unauthorized",
	}
}

// BadRequestError is error for the case of bad request
func BadRequestError(err error) error {
	return Error{
		Err:     err,
		Code:    BadRequestErrorCode,
		Message: err.Error(),
		Detail:  "bad request",
	}
}

// ForbiddenError is error for the case of forbidden
func ForbiddenError(err error) error {
	return Error{
		Err:     err,
		Code:    ForbiddenErrorCode,
		Message: err.Error(),
		Detail:  "forbidden",
	}
}

// PreconditionFailedError is error for the case of precondition failed
func PreconditionFailedError(err error) error {
	return Error{
		Err:     err,
		Code:    PreconditionErrorCode,
		Message: err.Error(),
		Detail:  "precondition failed",
	}
}

// UnknownError ...
func UnknownError(err error) error {
	return Error{
		Err:     err,
		Code:    GeneralErrorCode,
		Message: "unknown",
		Detail:  "generic error",
	}
}
