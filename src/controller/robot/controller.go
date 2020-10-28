package robot

import (
	"context"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/robot"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"time"
)

var (
	// RobotCtr is a global variable for the default robot account controller implementation
	RobotCtr = NewController()
)

// Controller to handle the requests related with robot account
type Controller interface {
	// GetRobotAccount ...
	GetRobotAccount(ctx context.Context, id int64) (*model.Robot, error)

	// CreateRobotAccount ...
	CreateRobotAccount(ctx context.Context, robotReq *model.RobotCreate) (*model.Robot, error)

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
}

// NewController ...
func NewController() Controller {
	return &DefaultAPIController{
		robotMgr: robot.Mgr,
		proMgr:   project.Mgr,
	}
}

// GetRobotAccount ...
func (d *DefaultAPIController) GetRobotAccount(ctx context.Context, id int64) (*model.Robot, error) {
	return d.robotMgr.GetRobotAccount(id)
}

// CreateRobotAccount ...
// 1, generate secret
// 2, call the
func (d *DefaultAPIController) CreateRobotAccount(ctx context.Context, robotReq *model.RobotCreate) (*model.Robot, error) {
	if robotReq.ExpiresAt == 0 {
		tokenDuration := time.Duration(config.RobotTokenDuration()) * time.Minute
		robotReq.ExpiresAt = time.Now().UTC().Add(tokenDuration).Unix()
	}
	pro, err := d.proMgr.Get(ctx, robotReq.Name)
	if err != nil {
		return nil, err
	}
	robot := &model.Robot{
		Name:        robotReq.Name,
		Description: robotReq.Description,
		ProjectID:   pro.ProjectID,
		ExpiresAt:   robotReq.ExpiresAt,
		Visible:     robotReq.Visible,
	}
	id, err := d.robotMgr.CreateRobotAccount(robot)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// DeleteRobotAccount ...
func (d *DefaultAPIController) DeleteRobotAccount(ctx context.Context, id int64) error {
	return d.robotMgr.DeleteRobotAccount(id)
}

// UpdateRobotAccount ...
func (d *DefaultAPIController) UpdateRobotAccount(ctx context.Context, r *model.Robot) error {
	return d.robotMgr.UpdateRobotAccount(r)
}

func (d *DefaultAPIController) ListRobotAccount(ctx context.Context, query *q.Query) ([]*model.Robot, error) {
	return d.robotMgr.ListRobotAccount(query)
}
