package robot

import (
	"github.com/goharbor/harbor/src/core/api"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
)

// RobotAPI ...
type RobotAPI struct {
	api.BaseController
	robotAPIImpl Interface
}

// Prepare ...
func (r *RobotAPI) Prepare() {
	if r.Ctx.Request.URL.Path == "/api/robots" {
		r.robotAPIImpl = &RobotSystemLevelAPI{}
	}

	r.robotAPIImpl.Prepare()
}

// Post ...
func (r *RobotAPI) Post() {
	r.robotAPIImpl.Post()
}

// List list all the robots of a project
func (r *RobotAPI) List() {
	r.robotAPIImpl.List()
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
		r.HandleBadRequest(fmt.Sprintf("invalid robot ID: %s", r.GetStringFromPath(":id")))
		return
	}

	if err := dao.DeleteRobot(id); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to delete robot %d: %v", id, err))
		return
	}
}
