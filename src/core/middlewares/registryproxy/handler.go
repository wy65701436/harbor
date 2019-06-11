package registryproxy

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type proxyHandler struct {
	handler http.Handler
}

func New(urls ...string) http.Handler {
	var registryURL string
	var err error
	if len(urls) > 1 {
		log.Errorf("the parm, urls should have only 0 or 1 elements")
		return nil
	}
	if len(urls) == 0 {
		registryURL, err = config.RegistryURL()
		if err != nil {
			log.Error(err)
			return nil
		}
	} else {
		registryURL = urls[0]
	}
	targetURL, err := url.Parse(registryURL)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &proxyHandler{
		handler: &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				director(targetURL, req)
			},
			ModifyResponse: modifyResponse,
		},
	}

}

// Overwrite the http requests
func director(target *url.URL, req *http.Request) {
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
}

// Modify the http response
func modifyResponse(res *http.Response) error {
	// Get the push success notification to record.
	// Needs to match PUT manifest
	if res.Request.Method == http.MethodPut && res.StatusCode == http.StatusCreated {
		match, _, _ := util.MatchManifestURL(res.Request)
		if match {
			log.Infof("response info ... %s", res.StatusCode)
			log.Infof("response header ... %s", res.Header)
			log.Infof("response Request.URL ... %s", res.Request.URL)
			data, err := ioutil.ReadAll(res.Body)
			log.Infof("response body ... %s", data)
			if err != nil {
				log.Infof("response body ... %s", data)
			}
		}
	}
	return nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func (ph proxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ph.handler.ServeHTTP(rw, req)
}
