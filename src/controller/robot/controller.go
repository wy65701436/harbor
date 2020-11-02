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
	"github.com/goharbor/harbor/src/pkg/permission/types"
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
	GetRobotAccount(ctx context.Context, id int64, option *Option) (*Robot, error)

	// CreateRobotAccount ...
	CreateRobotAccount(ctx context.Context, r *Robot) (*Robot, error)

	// DeleteRobotAccount ...
	DeleteRobotAccount(ctx context.Context, id int64) error

	// UpdateRobotAccount ...
	UpdateRobotAccount(ctx context.Context, r *Robot) error

	// ListRobotAccount ...
	ListRobotAccount(ctx context.Context, query *q.Query, option *Option) ([]*Robot, error)
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
	return d.populate(ctx, robot, option), nil
}

// CreateRobotAccount ...
func (d *DefaultAPIController) CreateRobotAccount(ctx context.Context, r *Robot) (*Robot, error) {
	// TODO add data validation
	if err := d.setProjectID(ctx, r); err != nil {
		return nil, err
	}

	if r.ExpiresAt == 0 {
		tokenDuration := time.Duration(config.RobotTokenDuration()) * time.Minute
		r.ExpiresAt = time.Now().UTC().Add(tokenDuration).Unix()
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
		Name:        r.Name,
		Description: r.Description,
		ProjectID:   r.ProjectID,
		ExpiresAt:   r.ExpiresAt,
		Secret:      secret,
	})
	if err != nil {
		return nil, err
	}
	r.ID = robotId
	r.Name = fmt.Sprintf("%s%s", common.RobotPrefix, r.Name)
	r.Secret = secret

	if err := d.createPermission(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
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
func (d *DefaultAPIController) UpdateRobotAccount(ctx context.Context, r *Robot) error {
	if r == nil {
		return errors.New("cannot update a nil robot").WithCode(errors.BadRequestCode)
	}
	if err := d.robotMgr.UpdateRobotAccount(&r.Robot); err != nil {
		return err
	}
	if err := d.setProjectID(ctx, r); err != nil {
		return err
	}
	if err := d.rbacMgr.DeletePermissionByRole(ctx, ROBOTTYPE, r.ID); err != nil {
		return err
	}
	if err := d.createPermission(ctx, r); err != nil {
		return err
	}
	return nil
}

// ListRobotAccount ...
func (d *DefaultAPIController) ListRobotAccount(ctx context.Context, query *q.Query, option *Option) ([]*Robot, error) {
	robots, err := d.robotMgr.ListRobotAccount(query)
	if err != nil {
		return nil, err
	}
	var robotAccounts []*Robot
	for _, r := range robots {
		robotAccounts = append(robotAccounts, d.populate(ctx, r, option))
	}
	return robotAccounts, nil
}

func (d *DefaultAPIController) createPermission(ctx context.Context, r *Robot) error {
	if r == nil {
		return nil
	}

	for _, per := range r.Permissions {
		policy := &rbac_model.RbacPolicy{}
		switch per.Kind {
		case LEVELSYSTEM:
			policy.Scope = SCOPESYSTEM
			if per.Namespace == ALLPROJECT {
				policy.Scope = SCOPEALLPROJECT
			}
		case LEVELPROJECT:
			policy.Scope = fmt.Sprintf("/project/%d", r.ProjectID)
		}
		for _, access := range per.Access {
			policy.Resource = access.Resource.String()
			policy.Action = access.Action.String()
			policy.Effect = access.Effect.String()

			policyID, err := d.rbacMgr.CreateRbacPolicy(ctx, policy)
			if err != nil {
				return err
			}

			_, err = d.rbacMgr.CreatePermission(ctx, &rbac_model.RolePermission{
				RoleType:     ROBOTTYPE,
				RoleID:       r.ID,
				RBACPolicyID: policyID,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *DefaultAPIController) setProjectID(ctx context.Context, r *Robot) error {
	if r == nil {
		return nil
	}

	var projectID int64
	switch r.Level {
	case LEVELSYSTEM:
		projectID = SYSTEMPROJECTID
	case LEVELPROJECT:
		pro, err := d.proMgr.Get(ctx, r.Permissions[0].Namespace)
		if err != nil {
			return err
		}
		projectID = pro.ProjectID
	default:
		return errors.New(nil).WithMessage("unknown robot account level").WithCode(errors.BadRequestCode)
	}
	r.ProjectID = projectID
	return nil
}

func (d *DefaultAPIController) populate(ctx context.Context, r *model.Robot, option *Option) *Robot {
	if r == nil {
		return nil
	}
	robot := &Robot{
		Robot: *r,
	}
	if r.ProjectID == SYSTEMPROJECTID {
		robot.Level = LEVELSYSTEM
	} else {
		robot.Level = LEVELPROJECT
	}
	if option == nil {
		return robot
	}
	if option.WithPermission {
		d.populatePermissions(ctx, robot)
	}
	return robot
}

func (d *DefaultAPIController) populatePermissions(ctx context.Context, r *Robot) {
	if r == nil {
		return
	}
	rolePermissions, err := d.rbacMgr.GetPermissionsByRole(ctx, ROBOTTYPE, r.ID)
	if err != nil {
		log.Errorf("failed to get permissions of robot %d: %v", r.ID, err)
		return
	}
	if len(rolePermissions) == 0 {
		return
	}

	// scope: accesses
	accessMap := make(map[string][]*types.Policy)

	// group by scope
	for _, rp := range rolePermissions {
		_, exist := accessMap[rp.Scope]
		if !exist {
			accessMap[rp.Scope] = []*types.Policy{&types.Policy{
				Resource: types.Resource(rp.Resource),
				Action:   types.Action(rp.Action),
				Effect:   types.Effect(rp.Effect),
			}}
		} else {
			accesses := accessMap[rp.Scope]
			accesses = append(accesses, &types.Policy{
				Resource: types.Resource(rp.Resource),
				Action:   types.Action(rp.Action),
				Effect:   types.Effect(rp.Effect),
			})
			accessMap[rp.Scope] = accesses
		}
	}

	var permissions []Permission
	for scope, accesses := range accessMap {
		p := Permission{}
		if scope == SCOPESYSTEM {
			p.Kind = LEVELSYSTEM
			p.Namespace = "/"
		} else if scope == SCOPEALLPROJECT {
			p.Kind = LEVELPROJECT
			p.Namespace = ALLPROJECT
		} else {
			p.Kind = LEVELPROJECT
			p.Namespace = "??"
		}
		p.Access = accesses
		permissions = append(permissions, p)
	}
	r.Permissions = permissions
}
