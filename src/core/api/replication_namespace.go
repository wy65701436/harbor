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

package api

import (
	"fmt"
	"github.com/goharbor/harbor/src/replication/ng"
	adp "github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/event"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// ReplicationNamespaceAPI handles the replication adapter requests
type ReplicationNamespaceAPI struct {
	BaseController
	registryID int64
}

// Prepare ...
func (r *ReplicationNamespaceAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsSysAdmin() {
		if !r.SecurityCtx.IsAuthenticated() {
			r.HandleUnauthorized()
			return
		}
		r.HandleForbidden(r.SecurityCtx.GetUsername())
		return
	}
	id, err := r.GetInt64FromPath(":id")
	if err != nil {
		r.HandleBadRequest("invalid registry ID")
		return
	}
	r.registryID = id
}

// Get ...
func (r *ReplicationNamespaceAPI) Get() {
	var registry *model.Registry
	var err error

	if r.registryID > 0 {
		registry, err = ng.RegistryMgr.Get(r.registryID)
		if err != nil {
			r.HandleInternalServerError(fmt.Sprintf("failed to get registry %d: %v", r.registryID, err))
			return
		}
	} else if r.registryID == 0 {
		registry = event.GetLocalRegistry()
	} else {
		r.HandleInternalServerError(fmt.Sprintf("failed to get registry %d: %v", r.registryID, err))
		return
	}

	if registry == nil {
		r.HandleNotFound(fmt.Sprintf("registry %d not found", r.registryID))
		return
	}

	if !adp.HasFactory(registry.Type) {
		r.HandleInternalServerError(fmt.Sprintf("no adapter factory found for %d", r.registryID))
		return
	}

	regFactory, err := adp.GetFactory(registry.Type)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("fail to get adapter factory %s", registry.Type))
		return
	}
	regAdapter, err := regFactory(registry)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("fail to get adapter %s", registry.Type))
		return
	}

	query := &model.NamespaceQuery{
		Name: r.GetString("name"),
	}
	npResults, err := regAdapter.ListNamespaces(query)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("fail to list namespaces %s %v", registry.Type, err))
		return
	}

	r.Data["json"] = npResults
	r.ServeJSON()
}
