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

package manifest

import (
	"errors"
	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/api/repository"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/internal"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/registry/error"
	"net/http"
	"net/http/httputil"
)

// NewHandler returns the handler to handler manifest requests
func NewHandler(proMgr project.Manager, proxy *httputil.ReverseProxy) http.Handler {
	return &handler{
		proMgr: proMgr,
		proxy:  proxy,
	}
}

type handler struct {
	proMgr project.Manager
	proxy  *httputil.ReverseProxy
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodHead:
		h.head(w, req)
	case http.MethodGet:
		h.get(w, req)
	case http.MethodDelete:
		h.delete(w, req)
	case http.MethodPut:
		h.put(w, req)
	}
}

// make sure the artifact exist before proxying the request to the backend registry
func (h *handler) head(w http.ResponseWriter, req *http.Request) {
	// TODO check the existence
	h.proxy.ServeHTTP(w, req)
}

// make sure the artifact exist before proxying the request to the backend registry
func (h *handler) get(w http.ResponseWriter, req *http.Request) {
	// TODO check the existence
	h.proxy.ServeHTTP(w, req)
}

func (h *handler) delete(w http.ResponseWriter, req *http.Request) {
	// TODO implement, just delete from database
}

func (h *handler) put(w http.ResponseWriter, req *http.Request) {
	mf, ok := middleware.ManifestInfoFromContext(req.Context())
	if !ok {
		error.Handle(w, req, errors.New("cannot get the manifest information from request context"))
		return
	}
	log.Info("1111111")
	// make sure the repository exist before pushing the manifest
	_, repositoryID, err := repository.Ctl.Ensure(req.Context(), mf.ProjectID, mf.Repository)
	if err != nil {
		error.Handle(w, req, err)
		return
	}

	buffer := internal.NewResponseBuffer(w)
	log.Info("1111111")
	// proxy the req to the backend docker registry
	h.proxy.ServeHTTP(buffer, req)
	if !buffer.Success() {
		if _, err := buffer.Flush(); err != nil {
			log.Errorf("failed to flush: %v", err)
		}
		return
	}
	log.Info("1111111")

	// When got the response from the backend docker registry, the manifest and
	// tag are both ready, so we don't need to handle the issue anymore:
	// https://github.com/docker/distribution/issues/2625

	var tags []string
	var dgt string
	if mf.Digest != "" {
		dgt = mf.Digest
	} else {
		dgt = buffer.Header().Get("Docker-Content-Digest")
		tags = append(tags, mf.Tag)
	}

	_, _, err = artifact.Ctl.Ensure(req.Context(), repositoryID, dgt, tags...)
	if err != nil {
		error.Handle(w, req, err)
		return
	}

	// flush the origin response from the docker registry to the underlying response writer
	if _, err := buffer.Flush(); err != nil {
		log.Errorf("failed to flush: %v", err)
	}

	// TODO fire event, add access log in the event handler
}
