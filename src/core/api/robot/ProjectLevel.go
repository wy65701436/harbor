package robot

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/dao"
	"fmt"
	"net/http"
	"strconv"
	"github.com/goharbor/harbor/src/core/api"
)

type RobotProjectLevelAPI struct {
	api.BaseController
	robot   *models.Robot
	project *models.Project
}

func (rp *RobotProjectLevelAPI) Post() {
	var robotReq models.RobotReq
	rp.DecodeJSONReq(&robotReq)

	robot := models.Robot{
		Name:        robotReq.Name,
		Description: robotReq.Description,
		ProjectID:   rp.project.ProjectID,
		// TODO: use token service to generate token per access information
		Token: "this is a placeholder",
	}

	id, err := dao.AddRobot(&robot)
	if err != nil {
		rp.HandleInternalServerError(fmt.Sprintf("failed to create robot account: %v", err))
	}

	rp.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

func (rp *RobotProjectLevelAPI) Prepare() {

}

func (rp *RobotProjectLevelAPI) List() {

}
