// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package admin

import (
	"encoding/json"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job"
	job_model "github.com/goharbor/harbor/src/common/job/models"
	common_models "github.com/goharbor/harbor/src/common/models"
	api_models "github.com/goharbor/harbor/src/core/api/models"
	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api"
	"fmt"
	"errors"
)

var statusMap = map[string]string{
	job.JobServiceStatusPending:   models.JobPending,
	job.JobServiceStatusRunning:   models.JobRunning,
	job.JobServiceStatusStopped:   models.JobStopped,
	job.JobServiceStatusCancelled: models.JobCanceled,
	job.JobServiceStatusError:     models.JobError,
	job.JobServiceStatusSuccess:   models.JobFinished,
	job.JobServiceStatusScheduled: models.JobScheduled,
}

// Handler handles request on /service/notifications/jobs/adminjob/*, which listens to the webhook of jobservice.
type Handler struct {
	api.BaseController
	id     int64
	UUID   string
	status string
	JobKind     string
	Cron string
}

// Prepare ...
func (h *Handler) Prepare() {
	var data job_model.JobStatusChange
	err := json.Unmarshal(h.Ctx.Input.CopyBody(1<<32), &data)
	if err != nil {
		log.Errorf("Failed to decode job status change, error: %v", err)
		h.Abort("200")
		return
	}
	h.UUID = data.JobID
	status, ok := statusMap[data.Status]
	if !ok || status == models.JobStopped {
		log.Infof("drop the job status update event: job id-%d, status-%s", h.UUID, status)
		h.Abort("200")
		return
	}
	h.JobKind = data.Metadata.JobKind
	h.status = status
	h.Cron = data.Metadata.CronSpec
}

// HandleAdminJob handles the webhook of admin jobs
func (h *Handler) HandleAdminJob() {
	log.Infof("received admin job status update event: job-%d, status-%s", h.UUID, h.status)

	jobs, err := dao.GetAdminJobs(&common_models.AdminJobQuery{
		UUID: h.UUID,
	})
	if err != nil {
		h.HandleInternalServerError(fmt.Sprintf("failed to get admin jobs: %v", err))
		return
	}
	if len(jobs) > 1 {
		h.HandleStatusPreconditionFailed(fmt.Sprintf("Get more than one job with same UUID: %s", h.UUID))
		return
	}

	var id int64
	// Add job for GCScheduler is to record the history of scheduled GC
	if len(jobs) == 0 || jobs[0].Kind == api_models.GCScheduler {

		cronStr, err := h.getCronStr(h.JobKind)
		if err != nil {
			h.HandleInternalServerError(fmt.Sprintf("%v", err))
			return
		}

		id, err = dao.AddAdminJob(&common_models.AdminJob{
			Name: common_job.ImageGC,
			Kind: h.JobKind,
			UUID: h.UUID,
			Cron: cronStr,
		})
		if err != nil {
			h.HandleInternalServerError(fmt.Sprintf("%v", err))
			return
		}
	}else {
		id = jobs[0].ID
	}

	if err := dao.UpdateAdminJobStatus(id, h.status); err != nil {
		log.Errorf("Failed to update job status, id: %d, status: %s", h.id, h.status)
		h.HandleInternalServerError(err.Error())
		return
	}
}

func (h *Handler) getCronStr(jobKind string) (string, error) {
	if jobKind == "Generic" {
		return "{\"type\":\"Manual\",\"weekday\":0,\"offtime\":0}", nil
	}
	if jobKind == "Periodic" {
		jobs, err := dao.GetAdminJobs(&common_models.AdminJobQuery{
			Name: common_job.ImageGC,
			Kind: api_models.GCScheduler,
		})
		if err != nil {
			return "", err
		}
		if len(jobs) > 1 {
			return "", fmt.Errorf("Get more than one gc scheduler.")
		}
		return jobs[0].Cron, nil
	}
	return "", errors.New("Unsupported job kind.")
}