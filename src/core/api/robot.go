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
	"net/http"
	"github.com/goharbor/harbor/src/common/models"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"strconv"
)

// RobotAPI ...
type RobotAPI struct {
	BaseController
	robot   *models.Robot
	project *models.Project
}

// Prepare ...
func (r *RobotAPI) Prepare() {
	r.BaseController.Prepare()

	if !r.SecurityCtx.IsAuthenticated() {
		r.HandleUnauthorized()
		return
	}

	pid, err := r.GetInt64FromPath(":pid")
	if err != nil || pid <= 0 {
		text := "invalid project ID: "
		if err != nil {
			text += err.Error()
		} else {
			text += fmt.Sprintf("%d", pid)
		}
		r.HandleBadRequest(text)
		return
	}
	project, err := r.ProjectMgr.Get(pid)
	if err != nil {
		r.ParseAndHandleError(fmt.Sprintf("failed to get project %d", pid), err)
		return
	}
	if project == nil {
		r.HandleNotFound(fmt.Sprintf("project %d not found", pid))
		return
	}
	r.project = project

}

// Post ...
func (r *RobotAPI) Post() {
	var robotReq models.RobotReq
	r.DecodeJSONReq(&robotReq)

	robot := models.Robot{
		Name:        robotReq.Name,
		Description: robotReq.Description,
		ProjectID:   r.project.ProjectID,
		// TODO: use token service to generate token per access information
		Token: "this is a placeholder",
	}

	id, err := dao.AddRobot(&robot)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to create robot account: %v", err))
	}

	r.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// List list all the robots of a project
func (r *RobotAPI) List() {
	query := models.RobotQuery{
		ProjectID: r.project.ProjectID,
	}
	robots, err := dao.ListRobots(&query)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get robots %v", err))
	}
	r.Data["json"] = robots
	r.ServeJSON()
}

// Get get robot by id
func (r *RobotAPI) Get() {
	id, err := r.GetInt64FromPath(":id")
	if err != nil || id <= 0 {
		r.HandleBadRequest(fmt.Sprintf("invalid robot ID: %s", r.GetStringFromPath(":id")))
		return
	}

	robot, err := dao.GetRobotByID(id)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get robot %d: %v", id, err))
		return
	}
	if robot == nil {
		r.HandleNotFound(fmt.Sprintf("robot %d not found", id))
		return
	}

	r.Data["json"] = robot
	r.ServeJSON()
}

// Put disable or enable a robot account
func (r *RobotAPI) Put() {
	var robotReq models.RobotReq
	r.DecodeJSONReq(&robotReq)

	id, err := r.GetInt64FromPath(":id")
	if err != nil || id <= 0 {
		r.HandleBadRequest(fmt.Sprintf("invalid robot ID: %s", r.GetStringFromPath(":id")))
		return
	}

	robot := models.Robot{
		ID:       id,
		Disabled: robotReq.Disabled,
	}

	err = dao.UpdateRobot(&robot)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to update robot: %v", err))
	}
}

// Delete delete robot by id
func (r *RobotAPI) Delete() {
	id, err := r.GetInt64FromPath(":id")
	if err != nil || id <= 0 {
		r.HandleBadRequest(fmt.Sprintf("invalid robot ID: %s", l.GetStringFromPath(":id")))
		return
	}

	if err := dao.DeleteRobot(id); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to delete robot %d: %v", id, err))
		return
	}
}
