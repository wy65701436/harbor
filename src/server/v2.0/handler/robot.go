package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
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

func (rAPI *robotAPI) CreateRobot(ctx context.Context, params operation.CreateRobotParams) middleware.Responder {
	if err := rAPI.RequireAuthenticated(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	robotAccount := &robot.Robot{
		Level: params.Robot.Level,
	}

	lib.JSONCopy(robotAccount.Robot, params.Robot)
	lib.JSONCopy(robotAccount.Permissions, params.Robot.Permissions)

	if err := rAPI.validate(params.Robot); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.requireAccess(ctx, params.Robot); err != nil {
		return rAPI.SendError(ctx, err)
	}

	created, err := rAPI.robotCtl.Create(ctx, robotAccount)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), created.ID)
	return operation.NewCreateRobotCreated().WithLocation(location).WithPayload(&models.RobotCreated{
		ID:           created.ID,
		Name:         created.Name,
		Secret:       created.Secret,
		CreationTime: strfmt.DateTime(created.CreationTime),
	})

}

func (rAPI *robotAPI) DeleteRobot(ctx context.Context, params operation.DeleteRobotParams) middleware.Responder {
	return nil
}

func (rAPI *robotAPI) GetRobot(ctx context.Context, params operation.GetRobotParams) middleware.Responder {
	return nil
}

func (rAPI *robotAPI) GetRobotByID(ctx context.Context, params operation.GetRobotByIDParams) middleware.Responder {
	return nil
}

func (rAPI *robotAPI) UpdateRobot(ctx context.Context, params operation.UpdateRobotParams) middleware.Responder {
	return nil
}

func (rAPI *robotAPI) requireAccess(ctx context.Context, r *models.Robot) error {
	if r.Level == robot.LEVELSYSTEM {
		if err := rAPI.RequireSysAdmin(ctx); err != nil {
			return err
		}
	} else if r.Level == robot.LEVELPROJECT {
		if err := rAPI.RequireProjectAccess(ctx, r.Permissions[0].Namespace, rbac.ActionCreate, rbac.ResourceRobot); err != nil {
			return err
		}
	}
	return nil
}

// more validation
func (rAPI *robotAPI) validate(r *models.Robot) error {
	if !isValidLevel(r.Level) {
		return errors.New(nil).WithMessage("bad request error level input").WithCode(errors.BadRequestCode)
	}

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

func isValidLevel(l string) bool {
	switch l {
	case
		robot.LEVELSYSTEM,
		robot.LEVELPROJECT:
		return true
	}
	return false
}
