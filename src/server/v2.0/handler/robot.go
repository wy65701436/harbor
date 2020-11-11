package handler

import (
	"github.com/goharbor/harbor/src/pkg/robot"
)

func newRobotAPI() *robotAPI {
	return &robotAPI{
		robotCtl: robot.RobotCtr,
	}
}

type robotAPI struct {
	BaseAPI
	robotCtl robot.Controller
}
