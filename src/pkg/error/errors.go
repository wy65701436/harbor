package error

import (
	"encoding/json"
	"errors"
	"strconv"
)

var (
	// ErrNotFound ...
	ErrNotFound = errors.New("not found")

	// ErrUnAuthorized ...
	ErrUnAuthorized = errors.New("unAuthorized")

	// ErrConflict ...
	ErrConflict = errors.New("conflict")

	// ErrForbidden ...
	ErrForbidden = errors.New("forbidden")

	// ErrBadRequest ...
	ErrBadRequest = errors.New("bad request")

	// ErrorPrecondition ...
	ErrorPrecondition = errors.New("precondition failed")

	// ErrUnknown ...
	ErrUnknown = errors.New("unknown")
)

// Code ...
type Code string

// M ...
type M string

// D ...
type D string

// Error ...
type Error struct {
	Err     error `json:"-"`
	Code    Code  `json:"code"`
	Message M     `json:"message"`
	Detail  D     `json:"detail,omitempty"`
}

// Error returns a human readable error.
func (e Error) Error() string {
	if data, err := json.Marshal(e); err == nil {
		return string(data)
	}
	return "{}"
}

// WithMessage ...
func (e Error) WithMessage(msg M) Error {
	return Error{
		Err:     e.Err,
		Code:    e.Code,
		Message: msg,
		Detail:  e.Detail,
	}
}

// Unwrap ...
func (e Error) Unwrap() error { return e.Err }

// E ...
func E(args ...interface{}) error {
	e := Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case error:
			e.Err = arg
		case string:
			e.Err = errors.New(arg)
		case Code:
			e.Code = arg
		case M:
			e.Message = arg
		case D:
			e.Detail = arg
		}
	}
	return e
}

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
			err = UnknownError(daErr).(Error).WithMessage(M(err.Error()))
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

// Es ...
func Es(err error) Errors {
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
	return E(ErrNotFound, Code(strconv.Itoa(ObjectNotFoundErrorCode)), M(err.Error()), D("not found"))
}

// ConflictError is error for the case of object conflict
func ConflictError(err error) error {
	return E(ErrConflict, Code(strconv.Itoa(ObjectConflictErrorCode)), M(err.Error()), D("conflict"))
}

// UnauthorizedError is error for the case of unauthorized accessing
func UnauthorizedError(err error) error {
	return E(ErrUnAuthorized, Code(strconv.Itoa(UnAuthorizedErrorCode)), M(err.Error()), D("unauthorized"))
}

// BadRequestError is error for the case of bad request
func BadRequestError(err error) error {
	return E(ErrBadRequest, Code(strconv.Itoa(BadRequestErrorCode)), M(err.Error()), D("bad request"))
}

// ForbiddenError is error for the case of forbidden
func ForbiddenError(err error) error {
	return E(ErrForbidden, Code(strconv.Itoa(ForbiddenErrorCode)), M(err.Error()), D("forbidden"))
}

// PreconditionFailedError is error for the case of precondition failed
func PreconditionFailedError(err error) error {
	return E(ErrorPrecondition, Code(strconv.Itoa(PreconditionErrorCode)), M(err.Error()), D("precondition failed"))
}

// UnknownError ...
func UnknownError(err error) error {
	return E(err, Code(strconv.Itoa(GeneralErrorCode)), M("unknown"), D("Generic error returned when the error happen."))
}
