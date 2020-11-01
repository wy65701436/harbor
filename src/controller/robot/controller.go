package robot

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/rbac"
	rbac_model "github.com/goharbor/harbor/src/pkg/rbac/model"
	"github.com/goharbor/harbor/src/pkg/robot"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"time"
)

var (
	// RobotCtr is a global variable for the default robot account controller implementation
	RobotCtr = NewController()
)

const (
	LEVELSYSTEM  = "system"
	LEVELPROJECT = "project"

	ALLPROJECT = "*"

	SCOPESYSTEM     = "/system"
	SCOPEALLPROJECT = "/project/*"

	ROBOTTYPE       = "robotaccount"
	SYSTEMPROJECTID = 0
)

// Controller to handle the requests related with robot account
type Controller interface {
	// GetRobotAccount ...
	GetRobotAccount(ctx context.Context, id int64) (*Robot, error)

	// CreateRobotAccount ...
	CreateRobotAccount(ctx context.Context, robot *Robot) (*Robot, error)

	// DeleteRobotAccount ...
	DeleteRobotAccount(ctx context.Context, id int64) error

	// UpdateRobotAccount ...
	UpdateRobotAccount(ctx context.Context, r *model.Robot) error

	// ListRobotAccount ...
	ListRobotAccount(ctx context.Context, query *q.Query) ([]*model.Robot, error)
}

// DefaultAPIController ...
type DefaultAPIController struct {
	robotMgr robot.Manager
	proMgr   project.Manager
	rbacMgr  rbac.Manager
}

// NewController ...
func NewController() Controller {
	return &DefaultAPIController{
		robotMgr: robot.Mgr,
		proMgr:   project.Mgr,
		rbacMgr:  rbac.Mgr,
	}
}

// GetRobotAccount ...
func (d *DefaultAPIController) GetRobotAccount(ctx context.Context, id int64, option *Option) (*Robot, error) {
	robot, err := d.robotMgr.GetRobotAccount(id)
	if err != nil {
		return nil, err
	}
	r := &Robot{
		Robot: *robot,
	}
	if r.ProjectID == SYSTEMPROJECTID {
		r.Level = LEVELSYSTEM
	} else {
		r.Level = LEVELPROJECT
	}
	if option == nil {
		return r, nil
	}
	if option.WithPermission {
		d.populatePermissions(ctx, r)
	}

	return r, nil
}

// CreateRobotAccount ...
func (d *DefaultAPIController) CreateRobotAccount(ctx context.Context, robot *Robot) (*Robot, error) {
	// add data validation
	var projectID int64
	switch robot.Level {
	case LEVELSYSTEM:
		projectID = SYSTEMPROJECTID
	case LEVELPROJECT:
		pro, err := d.proMgr.Get(ctx, robot.Permissions[0].Namespace)
		if err != nil {
			return nil, err
		}
		projectID = pro.ProjectID
	default:
		return nil, errors.New(nil).WithMessage("unknown robot account level").WithCode(errors.BadRequestCode)
	}

	if robot.ExpiresAt == 0 {
		tokenDuration := time.Duration(config.RobotTokenDuration()) * time.Minute
		robot.ExpiresAt = time.Now().UTC().Add(tokenDuration).Unix()
	}

	key, err := config.SecretKey()
	if err != nil {
		return nil, err
	}
	str := utils.GenerateRandomString()
	secret, err := utils.ReversibleEncrypt(str, key)
	if err != nil {
		return nil, err
	}

	robotId, err := d.robotMgr.CreateRobotAccount(&model.Robot{
		Name:        robot.Name,
		Description: robot.Description,
		ProjectID:   projectID,
		ExpiresAt:   robot.ExpiresAt,
		Secret:      secret,
	})
	if err != nil {
		return nil, err
	}

	for _, per := range robot.Permissions {
		policy := &rbac_model.RbacPolicy{}
		switch per.Kind {
		case LEVELSYSTEM:
			policy.Scope = SCOPESYSTEM
			if per.Namespace == ALLPROJECT {
				policy.Scope = SCOPEALLPROJECT
			}
		case LEVELPROJECT:
			policy.Scope = fmt.Sprintf("/project/%d", projectID)
		}
		for _, access := range per.Access {
			policy.Resource = access.Resource.String()
			policy.Action = access.Action.String()
			policy.Effect = access.Effect.String()

			policyID, err := d.rbacMgr.CreateRbacPolicy(ctx, policy)
			if err != nil {
				return nil, err
			}

			_, err = d.rbacMgr.CreatePermission(ctx, &rbac_model.RolePermission{
				RoleType:     ROBOTTYPE,
				RoleID:       robotId,
				RBACPolicyID: policyID,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	robot.Name = fmt.Sprintf("%s%s", common.RobotPrefix, robot.Name)
	robot.Secret = secret
	robot.ID = robotId
	return robot, nil
}

// DeleteRobotAccount ...
func (d *DefaultAPIController) DeleteRobotAccount(ctx context.Context, id int64) error {
	if err := d.robotMgr.DeleteRobotAccount(id); err != nil {
		return err
	}
	if err := d.rbacMgr.DeletePermissionByRole(ctx, ROBOTTYPE, id); err != nil {
		return err
	}
	return nil
}

// UpdateRobotAccount ...
func (d *DefaultAPIController) UpdateRobotAccount(ctx context.Context, r *model.Robot) error {
	return d.robotMgr.UpdateRobotAccount(r)
}

// ListRobotAccount ...
func (d *DefaultAPIController) ListRobotAccount(ctx context.Context, query *q.Query) ([]*model.Robot, error) {
	return d.robotMgr.ListRobotAccount(query)
}

func (d *DefaultAPIController) populatePermissions(ctx context.Context, r *Robot) {
	rolePermissions, err := d.rbacMgr.GetPermissionsByRole(ctx, ROBOTTYPE, r.ID)
	if err != nil {
		log.Errorf("failed to get permissions of robot %d: %v", r.ID, err)
		return
	}
	if len(rolePermissions) == 0 {
		return
	}

	for _, rp := range rolePermissions {
		p := Permission{}
		if rp.Scope == SCOPESYSTEM {
			p.Kind = LEVELSYSTEM
			p.Namespace = "/"
		} else if rp.Scope == SCOPEALLPROJECT {
			p.Kind = LEVELPROJECT
			p.Namespace = ALLPROJECT
		} else {
			p.Kind = LEVELPROJECT
			p.Namespace = "??"
		}
	}
}
