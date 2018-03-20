package middlewares

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/vmware/harbor/src/ui/config"
)

var proxy *httputil.ReverseProxy

type RegProxyHandler struct {
	next http.Handler
}

func (rh RegProxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var err error
	var registryURL string
	registryURL, err = config.RegistryURL()
	if err != nil {
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}
	targetURL, err := url.Parse(registryURL)
	if err != nil {
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}
	proxy = httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(rw, req)
}
