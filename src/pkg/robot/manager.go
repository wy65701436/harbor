package robot

import (
	"context"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/robot/dao"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

var (
	// Mgr is a global variable for the default robot account manager implementation
	Mgr = NewDefaultRobotAccountManager()
)

// Manager ...
type Manager interface {
	// GetRobotAccount ...
	GetRobotAccount(ctx context.Context, id int64) (*model.Robot, error)

	// CreateRobotAccount ...
	CreateRobotAccount(ctx context.Context, m *model.Robot) (int64, error)

	// DeleteRobotAccount ...
	DeleteRobotAccount(ctx context.Context, id int64) error

	// DeleteByProjectID ...
	DeleteByProjectID(ctx context.Context, projectID int64) error

	// UpdateRobotAccount ...
	UpdateRobotAccount(ctx context.Context, m *model.Robot) error

	// ListRobotAccount ...
	ListRobotAccount(ctx context.Context, query *q.Query) ([]*model.Robot, error)
}

type defaultRobotManager struct {
	dao dao.RobotAccountDao
}

// NewDefaultRobotAccountManager return a new instance of defaultRobotManager
func NewDefaultRobotAccountManager() Manager {
	return &defaultRobotManager{
		dao: dao.New(),
	}
}

// GetRobotAccount ...
func (drm *defaultRobotManager) GetRobotAccount(ctx context.Context, id int64) (*model.Robot, error) {
	return drm.dao.GetRobotAccount(id)
}

// CreateRobotAccount ...
func (drm *defaultRobotManager) CreateRobotAccount(ctx context.Context, r *model.Robot) (int64, error) {
	return drm.dao.CreateRobotAccount(r)
}

// DeleteRobotAccount ...
func (drm *defaultRobotManager) DeleteRobotAccount(ctx context.Context, id int64) error {
	return drm.dao.DeleteRobotAccount(id)
}

// DeleteByProjectID ...
func (drm *defaultRobotManager) DeleteByProjectID(ctx context.Context, projectID int64) error {
	return drm.dao.DeleteByProjectID(ctx, projectID)
}

// UpdateRobotAccount ...
func (drm *defaultRobotManager) UpdateRobotAccount(ctx context.Context, r *model.Robot) error {
	return drm.dao.UpdateRobotAccount(r)
}

// ListRobotAccount ...
func (drm *defaultRobotManager) ListRobotAccount(ctx context.Context, query *q.Query) ([]*model.Robot, error) {
	return drm.dao.ListRobotAccounts(query)
}
