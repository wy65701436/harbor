// Copyright 2018 Project Harbor Authors
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

// Package utils contains methods to support security, cache, and webhook functions.
package utils

import (
	"net/http"
	"os"
	"time"

	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/core/service/token"
	"net/http/httptest"
)

// NewRepositoryClientForUI creates a repository client that can only be used to
// access the internal registry
func NewRepositoryClientForUI(username, repository string) (*registry.Repository, error) {
	endpoint, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}
	return newRepositoryClient(endpoint, username, repository)
}

// NewRepositoryClientForLocal creates a repository client that can only be used to
// access the internal registry with 127.0.0.1
func NewRepositoryClientForLocal(username, repository string) (*registry.Repository, error) {
	// The 127.0.0.1:8080 is not reachable as we do not enable core in UT env.
	if os.Getenv("UTTEST") == "true" {
		return NewRepositoryClientForUI(username, repository)
	}
	return newRepositoryClient(config.LocalCoreURL(), username, repository)
}

func newRepositoryClient(endpoint, username, repository string) (*registry.Repository, error) {
	uam := &auth.UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}
	authorizer := auth.NewRawTokenAuthorizer(username, token.Registry)
	transport := registry.NewTransport(NewTransportWithMiddleware(http.DefaultTransport), authorizer, uam)
	client := &http.Client{
		Transport: transport,
	}
	return registry.NewRepository(repository, endpoint, client)
}

type TransportWithMiddleware struct {
	transport http.RoundTripper
}

// NewTransportWithMiddleware ...
func NewTransportWithMiddleware(transport http.RoundTripper) *TransportWithMiddleware {
	return &TransportWithMiddleware{
		transport: transport,
	}
}

// RoundTrip ...
func (t *TransportWithMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {

	// Do "before sending requests" actions here.
	fmt.Printf("Sending request to %v\n", req.URL)
	handlerChain := middlewares.New(middlewares.Middlewares).Create()
	head := handlerChain.Then(nil)
	rw := httptest.NewRecorder()
	customResW := util.NewCustomResponseWriter(rw)
	head.ServeHTTP(customResW, req)

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	log.Debugf("%d | %s %s", resp.StatusCode, req.Method, req.URL.String())

	return resp, err
}

// WaitForManifestReady implements exponential sleeep to wait until manifest is ready in registry.
// This is a workaround for https://github.com/docker/distribution/issues/2625
func WaitForManifestReady(repository string, tag string, maxRetry int) bool {
	// The initial wait interval, hard-coded to 50ms
	interval := 50 * time.Millisecond
	repoClient, err := NewRepositoryClientForUI("harbor-core", repository)
	if err != nil {
		log.Errorf("Failed to create repo client.")
		return false
	}
	for i := 0; i < maxRetry; i++ {
		_, exist, err := repoClient.ManifestExist(tag)
		if err != nil {
			log.Errorf("Unexpected error when checking manifest existence, image:  %s:%s, error: %v", repository, tag, err)
			continue
		}
		if exist {
			return true
		}
		log.Warningf("manifest for image %s:%s is not ready, retry after %v", repository, tag, interval)
		time.Sleep(interval)
		interval = interval * 2
	}
	return false
}
