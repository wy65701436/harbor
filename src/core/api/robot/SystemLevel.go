package robot

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/api"
)

type RobotSystemLevelAPI struct {
	api.BaseController
	robot   *models.Robot
	project *models.Project
}

func (rp *RobotSystemLevelAPI) Post() {}

func (rp *RobotSystemLevelAPI) Prepare() {}

func (rp *RobotSystemLevelAPI) List() {}
