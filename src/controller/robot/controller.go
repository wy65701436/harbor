package robot

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/rbac"
	rbac_model "github.com/goharbor/harbor/src/pkg/rbac/model"
	robot "github.com/goharbor/harbor/src/pkg/robot2"
	"github.com/goharbor/harbor/src/pkg/robot2/model"
	"time"
)

var (
	// Ctr is a global variable for the default robot account controller implementation
	Ctr = NewController()
)

// Controller to handle the requests related with robot account
type Controller interface {
	// Get ...
	Get(ctx context.Context, id int64, option *Option) (*Robot, error)

	// Create ...
	Create(ctx context.Context, r *Robot) (*Robot, error)

	// Delete ...
	Delete(ctx context.Context, id int64) error

	// Update ...
	Update(ctx context.Context, r *Robot) error

	// List ...
	List(ctx context.Context, query *q.Query, option *Option) ([]*Robot, error)
}

// controller ...
type controller struct {
	robotMgr robot.Manager
	proMgr   project.Manager
	rbacMgr  rbac.Manager
}

// NewController ...
func NewController() Controller {
	return &controller{
		robotMgr: robot.Mgr,
		proMgr:   project.Mgr,
		rbacMgr:  rbac.Mgr,
	}
}

// Get ...
func (d *controller) Get(ctx context.Context, id int64, option *Option) (*Robot, error) {
	robot, err := d.robotMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return d.populate(ctx, robot, option), nil
}

// Create ...
func (d *controller) Create(ctx context.Context, r *Robot) (*Robot, error) {
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

	robotId, err := d.robotMgr.Create(ctx, &model.Robot{
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
	r.Name = fmt.Sprintf("%s%s", config.RobotPrefix(), r.Name)
	r.Secret = secret

	if err := d.createPermission(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Delete ...
func (d *controller) Delete(ctx context.Context, id int64) error {
	if err := d.robotMgr.Delete(ctx, id); err != nil {
		return err
	}
	if err := d.rbacMgr.DeletePermissionByRole(ctx, ROBOTTYPE, id); err != nil {
		return err
	}
	return nil
}

// Update ...
func (d *controller) Update(ctx context.Context, r *Robot) error {
	if r == nil {
		return errors.New("cannot update a nil robot").WithCode(errors.BadRequestCode)
	}
	if err := d.robotMgr.Update(ctx, &r.Robot); err != nil {
		return err
	}
	if err := d.setProjectID(ctx, r); err != nil {
		return err
	}
	// update the permission
	if err := d.rbacMgr.DeletePermissionByRole(ctx, ROBOTTYPE, r.ID); err != nil {
		return err
	}
	if err := d.createPermission(ctx, r); err != nil {
		return err
	}
	return nil
}

// List ...
func (d *controller) List(ctx context.Context, query *q.Query, option *Option) ([]*Robot, error) {
	robots, err := d.robotMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var robotAccounts []*Robot
	for _, r := range robots {
		robotAccounts = append(robotAccounts, d.populate(ctx, r, option))
	}
	return robotAccounts, nil
}

func (d *controller) createPermission(ctx context.Context, r *Robot) error {
	if r == nil {
		return nil
	}

	for _, per := range r.Permissions {
		policy := &rbac_model.RbacPolicy{}
		scope, err := per.toScope(r.ProjectID)
		if err != nil {
			return err
		}
		policy.Scope = scope

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

func (d *controller) populate(ctx context.Context, r *model.Robot, option *Option) *Robot {
	if r == nil {
		return nil
	}
	robot := &Robot{
		Robot: *r,
	}
	robot.setLevel()
	if option == nil {
		return robot
	}
	if option.WithPermission {
		d.populatePermissions(ctx, robot)
	}
	return robot
}

func (d *controller) populatePermissions(ctx context.Context, r *Robot) {
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

	var permissions []*Permission
	for scope, accesses := range accessMap {
		p := &Permission{}
		kind, namespace, err := d.decodeScope(ctx, scope, r.ProjectID)
		if err != nil {
			log.Errorf("failed to decode scope of robot %d: %v", r.ID, err)
			continue
		}
		p.Kind = kind
		p.Namespace = namespace
		p.Access = accesses
		permissions = append(permissions, p)
	}
	r.Permissions = permissions
}

func (d *controller) setProjectID(ctx context.Context, r *Robot) error {
	if r == nil {
		return nil
	}
	var projectID int64
	switch r.Level {
	case LEVELSYSTEM:
		projectID = 0
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

func (d *controller) decodeScope(ctx context.Context, scope string, projectID int64) (kind, namespace string, err error) {
	if scope == "" {
		return
	}
	if scope == SCOPESYSTEM {
		kind = LEVELSYSTEM
		namespace = "/"
	} else if scope == SCOPEALLPROJECT {
		kind = LEVELPROJECT
		namespace = "*"
	} else {
		kind = LEVELPROJECT
		pro, err := d.proMgr.Get(ctx, projectID)
		if err != nil {
			return "", "", err
		}
		namespace = pro.Name
	}
	return
}
