// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/config"
)

var proxy = newProxy()

// newProxy creates a reverse proxy to the registry service.
// It uses a custom Director function instead of NewSingleHostReverseProxy
// to properly rewrite the Host header, which is required for service mesh
// compatibility (e.g., Istio routes based on Host header, not IP).
// See: https://pkg.go.dev/net/http/httputil#NewSingleHostReverseProxy
// See: https://istio.io/latest/docs/ops/configuration/traffic-management/traffic-routing/#http
func newProxy() http.Handler {
	regURL, _ := config.RegistryURL()
	targetURL, err := url.Parse(regURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse the URL of registry: %v", err))
	}

	director := func(req *http.Request) {
		// Rewrite the request URL to point to the target registry
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = singleJoiningSlash(targetURL.Path, req.URL.Path)

		// Rewrite the Host header to match the target registry.
		// This is critical for service mesh routing (e.g., Istio).
		req.Host = targetURL.Host

		// Merge query parameters
		if targetURL.RawQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetURL.RawQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetURL.RawQuery + "&" + req.URL.RawQuery
		}

		// Add basic authentication for registry access
		u, p := config.RegistryCredential()
		if u != "" && p != "" {
			req.SetBasicAuth(u, p)
		}
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
	}

	if commonhttp.InternalTLSEnabled() {
		proxy.Transport = commonhttp.GetHTTPTransport()
	}

	return proxy
}

// singleJoiningSlash joins two URL paths with a single slash.
// This is the same logic used by httputil.NewSingleHostReverseProxy.
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
