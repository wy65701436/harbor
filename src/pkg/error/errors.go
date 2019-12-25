package error

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Error ... without detail...
type Error struct {
	Cause   error  `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns a human readable error.
func (e *Error) Error() string {
	return fmt.Sprintf("%v, %s, %s", e.Cause, e.Code, e.Message)
}

// WithMessage ...
func (e *Error) WithMessage(msg string) *Error {
	e.Message = msg
	return e
}

// WithCode
func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

// Unwrap ...
func (e *Error) Unwrap() error { return e.Cause }

// Errors ...
type Errors []error

var _ error = Errors{}

// MarshalJSON converts slice of error
func (errs Errors) Error() string {
	var tmpErrs struct {
		Errors []Error `json:"errors,omitempty"`
	}

	for _, e := range errs {
		var err error
		switch e.(type) {
		case *Error:
			err = e.(*Error)
		default:
			err = UnknownError(e).WithMessage(err.Error())
		}
		tmpErrs.Errors = append(tmpErrs.Errors, Error{
			Cause:   err.(*Error).Cause,
			Code:    err.(*Error).Code,
			Message: err.(*Error).Message,
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
	// ObjectNotFoundCode is code for the error of no object found
	ObjectNotFoundCode = "NOT_FOUND"
	// ObjectConflictErrorCode ...
	ObjectConflictErrorCode = "CONFLICT"
	// UnAuthorizedErrorCode ...
	UnAuthorizedErrorCode = "UNAUTHORIZED"
	// BadRequestErrorCode ...
	BadRequestErrorCode = "BAD_REQUEST"
	// ForbiddenErrorCode ...
	ForbiddenErrorCode = "FORBIDDER"
	// PreconditionErrorCode ...
	PreconditionErrorCode = "PRECONDITION"
	// GeneralErrorCode ...
	GeneralErrorCode = "UNKNOWN"
)

// NewErr ...
func NewErr(err error) *Error {
	if _, ok := err.(*Error); ok {
		err = err.(*Error).Unwrap()
	}
	return &Error{
		Cause: err,
	}
}

// NotFoundError is error for the case of object not found
func NotFoundError(err error) *Error {
	return NewErr(err).WithCode(ObjectNotFoundCode).WithMessage("resource not found")
}

// ConflictError is error for the case of object conflict
func ConflictError(err error) *Error {
	return NewErr(err).WithCode(ObjectConflictErrorCode).WithMessage("resource conflict")
}

// UnauthorizedError is error for the case of unauthorized accessing
func UnauthorizedError(err error) *Error {
	return NewErr(err).WithCode(UnAuthorizedErrorCode).WithMessage("unauthorized")
}

// BadRequestError is error for the case of bad request
func BadRequestError(err error) *Error {
	return NewErr(err).WithCode(BadRequestErrorCode).WithMessage("bad request")
}

// ForbiddenError is error for the case of forbidden
func ForbiddenError(err error) *Error {
	return NewErr(err).WithCode(ForbiddenErrorCode).WithMessage("forbidden")
}

// PreconditionFailedError is error for the case of precondition failed
func PreconditionFailedError(err error) *Error {
	return NewErr(err).WithCode(PreconditionErrorCode).WithMessage("preconfition")
}

// UnknownError ...
func UnknownError(err error) *Error {
	return NewErr(err).WithCode(GeneralErrorCode).WithMessage("unknown")
}

// IsErr
func IsErr(err error, code string) bool {
	_, ok := err.(*Error)
	if !ok {
		return false
	}
	return strings.Compare(err.(*Error).Code, code) == 0
}
