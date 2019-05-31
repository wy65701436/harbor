package middlewares

import (
	"github.com/goharbor/harbor/src/core/middlewares/registryproxy"
	"net/http"
)

var head http.Handler

// Init initialize the Proxy instance and handler chain.
func Init() error {
	var err error
	var ph http.Handler
	ph, err = registryproxy.New()
	if err != nil {
		return err
	}
	handlerChain := New(Middlewares).CreateChain()
	head, err = handlerChain.Then(ph)
	if err != nil {
		return err
	}
	return nil
}

// Handle handles the request.
func Handle(rw http.ResponseWriter, req *http.Request) {
	head.ServeHTTP(rw, req)
}
