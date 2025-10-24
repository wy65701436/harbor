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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

func TestSingleJoiningSlash(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{
			name:     "both have slashes",
			a:        "http://registry/",
			b:        "/v2/test",
			expected: "http://registry/v2/test",
		},
		{
			name:     "neither has slashes",
			a:        "http://registry",
			b:        "v2/test",
			expected: "http://registry/v2/test",
		},
		{
			name:     "only a has slash",
			a:        "http://registry/",
			b:        "v2/test",
			expected: "http://registry/v2/test",
		},
		{
			name:     "only b has slash",
			a:        "http://registry",
			b:        "/v2/test",
			expected: "http://registry/v2/test",
		},
		{
			name:     "empty b",
			a:        "http://registry/",
			b:        "",
			expected: "http://registry/",
		},
		{
			name:     "empty a",
			a:        "",
			b:        "/v2/test",
			expected: "/v2/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := singleJoiningSlash(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProxyHostHeaderRewrite(t *testing.T) {
	// Create a test backend server that echoes back the Host header
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Forwarded-Host", r.Host)
		w.Header().Set("X-Request-URL", r.URL.String())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	// Set up config to point to our test backend
	config.DefaultMgr().Set(nil, "registry_url", backend.URL)

	// Parse backend URL for verification
	backendURL, err := url.Parse(backend.URL)
	require.NoError(t, err)
	config.DefaultMgr().Set(nil, "registry_username", "testuser")
	config.DefaultMgr().Set(nil, "registry_password", "testpass")

	// Create the proxy
	proxyHandler := newProxy()

	// Create a test request
	req := httptest.NewRequest("GET", "http://harbor.example.com/v2/test/manifests/latest", nil)
	req.Host = "harbor.example.com"

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the request through the proxy
	proxyHandler.ServeHTTP(rr, req)

	// Verify the Host header was rewritten to the backend host
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, backendURL.Host, rr.Header().Get("X-Forwarded-Host"),
		"Host header should be rewritten to registry URL for Istio compatibility")
}

func TestProxyBasicAuth(t *testing.T) {
	// Create a test backend server that checks for basic auth
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("No auth"))
			return
		}
		w.Header().Set("X-Auth-User", username)
		w.Header().Set("X-Auth-Pass", password)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	// Set up config
	config.DefaultMgr().Set(nil, "registry_url", backend.URL)
	config.DefaultMgr().Set(nil, "registry_username", "harbor_registry_user")
	config.DefaultMgr().Set(nil, "registry_password", "harbor_registry_password")

	// Create the proxy
	proxyHandler := newProxy()

	// Create a test request
	req := httptest.NewRequest("GET", "http://harbor.example.com/v2/_catalog", nil)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the request through the proxy
	proxyHandler.ServeHTTP(rr, req)

	// Verify basic auth was added
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "harbor_registry_user", rr.Header().Get("X-Auth-User"))
	assert.Equal(t, "harbor_registry_password", rr.Header().Get("X-Auth-Pass"))
}

func TestProxyPathRewrite(t *testing.T) {
	tests := []struct {
		name         string
		registryURL  string
		requestPath  string
		expectedPath string
	}{
		{
			name:         "simple path",
			registryURL:  "http://registry:5000",
			requestPath:  "/v2/test/manifests/latest",
			expectedPath: "/v2/test/manifests/latest",
		},
		{
			name:         "registry with path",
			registryURL:  "http://registry:5000/registry",
			requestPath:  "/v2/test/manifests/latest",
			expectedPath: "/registry/v2/test/manifests/latest",
		},
		{
			name:         "registry with trailing slash",
			registryURL:  "http://registry:5000/registry/",
			requestPath:  "/v2/test/manifests/latest",
			expectedPath: "/registry/v2/test/manifests/latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test backend server that echoes back the path
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-Path", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			}))
			defer backend.Close()

			// Parse backend URL and append the registry path
			backendURL, err := url.Parse(backend.URL)
			require.NoError(t, err)

			registryURL, err := url.Parse(tt.registryURL)
			require.NoError(t, err)

			fullURL := backend.URL + registryURL.Path

			// Set up config
			config.DefaultMgr().Set(nil, "registry_url", fullURL)
			config.DefaultMgr().Set(nil, "registry_username", "test")
			config.DefaultMgr().Set(nil, "registry_password", "test")

			// Create the proxy
			proxyHandler := newProxy()

			// Create a test request
			req := httptest.NewRequest("GET", "http://harbor.example.com"+tt.requestPath, nil)

			// Record the response
			rr := httptest.NewRecorder()

			// Serve the request through the proxy
			proxyHandler.ServeHTTP(rr, req)

			// Verify the path was correctly rewritten
			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tt.expectedPath, rr.Header().Get("X-Request-Path"))
		})
	}
}

func TestProxyQueryParameters(t *testing.T) {
	// Create a test backend server that echoes back query parameters
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Query", r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	tests := []struct {
		name          string
		registryQuery string
		requestQuery  string
		expectedQuery string
	}{
		{
			name:          "no queries",
			registryQuery: "",
			requestQuery:  "",
			expectedQuery: "",
		},
		{
			name:          "only request query",
			registryQuery: "",
			requestQuery:  "n=10",
			expectedQuery: "n=10",
		},
		{
			name:          "only registry query",
			registryQuery: "service=registry",
			requestQuery:  "",
			expectedQuery: "service=registry",
		},
		{
			name:          "both queries",
			registryQuery: "service=registry",
			requestQuery:  "n=10",
			expectedQuery: "service=registry&n=10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registryURL := backend.URL
			if tt.registryQuery != "" {
				registryURL += "?" + tt.registryQuery
			}

			// Set up config
			config.DefaultMgr().Set(nil, "registry_url", registryURL)
			config.DefaultMgr().Set(nil, "registry_username", "test")
			config.DefaultMgr().Set(nil, "registry_password", "test")

			// Create the proxy
			proxyHandler := newProxy()

			// Create a test request
			requestURL := "http://harbor.example.com/v2/_catalog"
			if tt.requestQuery != "" {
				requestURL += "?" + tt.requestQuery
			}
			req := httptest.NewRequest("GET", requestURL, nil)

			// Record the response
			rr := httptest.NewRecorder()

			// Serve the request through the proxy
			proxyHandler.ServeHTTP(rr, req)

			// Verify query parameters were correctly merged
			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tt.expectedQuery, rr.Header().Get("X-Query"))
		})
	}
}
