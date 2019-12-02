package errors

import (
	"net/http"
)

const errorGroup = "harbor.api.v2"

var (
	ErrUnAuthorized = Register(errorGroup, ErrorDescriptor{
		Value:          "UnAuthorized",
		Message:        "UnAuthorized",
		Description:    "No valid authorized information associate with the request.",
		HTTPStatusCode: http.StatusUnauthorized,
	})
	ErrInvalidGeneral = Register(errorGroup, ErrorDescriptor{
		Value:          "INVALID_GENERAL",
		Message:        "the provided request has invalid attribute.",
		Description:    "Invalid information associate with the request.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrInvalidPID = Register(errorGroup, ErrorDescriptor{
		Value:          "INVALID_PID",
		Message:        "the provided project ID is not valid.",
		Description:    "Invalid project ID associate with the request.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrConflict = Register(errorGroup, ErrorDescriptor{
		Value:          "RESOURCE_CONFLICT",
		Message:        "The resource requested is conflict.",
		Description:    "Resource conflicted in the backend.",
		HTTPStatusCode: http.StatusConflict,
	})
	ErrNotFound = Register(errorGroup, ErrorDescriptor{
		Value:          "NO_RESOURCE_FOUND",
		Message:        "The resource requested is not found.",
		Description:    "No resource found in the backend",
		HTTPStatusCode: http.StatusNotFound,
	})
	ErrorUnknown = Register(errorGroup, ErrorDescriptor{
		Value:          "UNKNOWN",
		Message:        "unknown error",
		Description:    `Generic error returned when the error does not have an API classification.`,
		HTTPStatusCode: http.StatusInternalServerError,
	})
)
