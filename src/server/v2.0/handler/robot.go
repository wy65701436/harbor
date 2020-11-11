package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/robot"
	"strings"
)

func newRobotAPI() *robotAPI {
	return &robotAPI{
		robotCtl: robot.Ctr,
	}
}

type robotAPI struct {
	BaseAPI
	robotCtl robot.Controller
}

func (api *robotAPI) CreateRobot(ctx context.Context, params operation.CreateRobotParams) middleware.Responder {
	if err := api.RequireAuthenticated(ctx); err != nil {
		return api.SendError(ctx, err)
	}

	robotAccount := &robot.Robot{
		Level: params.Robot.Level,
	}

	lib.JSONCopy(robotAccount.Robot, params.Robot)
	lib.JSONCopy(robotAccount.Permissions, params.Robot.Permissions)

	if err := api.validate(params.Robot); err != nil {
		return api.SendError(ctx, err)
	}

	if err := api.requireAccess(ctx, params.Robot); err != nil {
		return api.SendError(ctx, err)
	}

	created, err := api.robotCtl.Create(ctx, robotAccount)
	if err != nil {
		return api.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), created.ID)
	return operation.NewCreateRobotCreated().WithLocation(location)

}

func (api *robotAPI) requireAccess(ctx context.Context, r *models.Robot) error {
	if r.Level == robot.LEVELSYSTEM {
		if err := api.RequireSysAdmin(ctx); err != nil {
			return err
		}
	} else if r.Level == robot.LEVELPROJECT {
		if err := api.RequireProjectAccess(ctx, r.Permissions[0].Namespace, rbac.ActionCreate, rbac.ResourceRobot); err != nil {
			return err
		}
	}
	return nil
}

// more validation
func (api *robotAPI) validate(r *models.Robot) error {
	if len(r.Permissions) == 0 {
		return errors.New(nil).WithMessage("bad request empty permission").WithCode(errors.BadRequestCode)
	}

	if r.Level == robot.LEVELPROJECT {
		// to create a project robot, the permission must be only one project scope.
		if len(r.Permissions) > 1 {
			return errors.New(nil).WithMessage("bad request permission").WithCode(errors.BadRequestCode)
		}
	}
	return nil
}
