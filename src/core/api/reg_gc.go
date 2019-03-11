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
	"net/http"
	"os"
	"strconv"

	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/core/api/models"
)

// GCAPI handles request of harbor admin...
type GCAPI struct {
	AJAPI
}

// Prepare validates the URL and parms, it needs the system admin permission.
func (gc *GCAPI) Prepare() {
	gc.BaseController.Prepare()
	if !gc.SecurityCtx.IsAuthenticated() {
		gc.HandleUnauthorized()
		return
	}
	if !gc.SecurityCtx.IsSysAdmin() {
		gc.HandleForbidden(gc.SecurityCtx.GetUsername())
		return
	}
}

// Post ...
func (gc *GCAPI) Post() {
	ajr := models.AdminJobReq{}
	gc.DecodeJSONReqAndValidate(&ajr)
	ajr.Name = common_job.ImageGC
	ajr.Parameters = map[string]interface{}{
		"redis_url_reg": os.Getenv("_REDIS_URL_REG"),
	}
	gc.submitAdminJob(&ajr)
	gc.Redirect(http.StatusCreated, strconv.FormatInt(ajr.ID, 10))
}

// Put ...
func (gc *GCAPI) Put() {
	ajr := models.AdminJobReq{}
	gc.DecodeJSONReqAndValidate(&ajr)
	ajr.Name = common_job.ImageGC
	gc.updateAdminSchedule(ajr)
}

// GetGC ...
func (gc *GCAPI) GetGC() {
	id, err := gc.GetInt64FromPath(":id")
	if err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("need to specify gc id"))
		return
	}
	gc.getAdminJob(id)
}

// List ...
func (gc *GCAPI) List() {
	gc.listAdminJobs(common_job.ImageGC)
}

// Get gets GC schedule ...
func (gc *GCAPI) Get() {
	gc.getAdminSchedule(common_job.ImageGC)
}

// GetLog ...
func (gc *GCAPI) GetLog() {
	id, err := gc.GetInt64FromPath(":id")
	if err != nil {
		gc.HandleBadRequest("invalid ID")
		return
	}
	gc.getAdminJobLog(id)
}
