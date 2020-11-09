package robot

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot2/model"
)

const (
	LEVELSYSTEM  = "system"
	LEVELPROJECT = "project"

	SCOPESYSTEM     = "/system"
	SCOPEALLPROJECT = "/project/*"

	ROBOTTYPE = "robotaccount"
)

// Robot ...
type Robot struct {
	model.Robot
	ProjectName string
	Level       string
	Permissions []*Permission `json:"permissions"`
}

// setLevel, 0 is a system level robot, others are project level.
func (r *Robot) setLevel() {
	if r.ProjectID == 0 {
		r.Level = LEVELSYSTEM
	} else {
		r.Level = LEVELPROJECT
	}
}

// Permission ...
type Permission struct {
	Kind      string          `json:"kind"`
	Namespace string          `json:"namespace"`
	Access    []*types.Policy `json:"access"`
}

// toScope is to translate the permission kind and namespace to scope.
func (p *Permission) toScope(projectID int64) (string, error) {
	switch p.Kind {
	case LEVELSYSTEM:
		return SCOPESYSTEM, nil
		if p.Namespace == "*" {
			return SCOPEALLPROJECT, nil
		}
	case LEVELPROJECT:
		return fmt.Sprintf("/project/%d", projectID), nil
	}
	return "", errors.New(nil).WithMessage("unknown robot kind").WithCode(errors.BadRequestCode)
}

// Option ...
type Option struct {
	WithPermission bool
}
