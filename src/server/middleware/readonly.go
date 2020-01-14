package middleware

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	internal_errors "github.com/goharbor/harbor/src/internal/error"
	"net/http"
)

type readonlyHandler struct {
	next http.Handler
}

// ReadOnly middleware reject request when harbor set to readonly
func ReadOnly() func(http.Handler) http.Handler {
	return new
}

// new ...
func new(next http.Handler) http.Handler {
	return &readonlyHandler{
		next: next,
	}
}

// ServeHTTP ...
// it should be applied into http.MethodDelete, http.MethodPost, http.MethodPatch, http.MethodPut
func (rh readonlyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if config.ReadOnly() {
		log.Warningf("The request is prohibited in readonly mode, url is: %s", req.URL.Path)
		pkgE := internal_errors.New(nil).WithCode("DENIED").WithMessage("The system is in read only mode. Any modification is prohibited.")
		http.Error(rw, internal_errors.NewErrs(pkgE).Error(), http.StatusForbidden)
		return
	}
	rh.next.ServeHTTP(rw, req)
}
