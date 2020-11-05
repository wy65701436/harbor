package robot

import (
	"fmt"
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

type Permission struct {
	Kind      string          `json:"kind"`
	Namespace string          `json:"namespace"`
	Access    []*types.Policy `json:"access"`
}

func (p *Permission) toScope(projectID int64) string {
	switch p.Kind {
	case LEVELSYSTEM:
		return SCOPESYSTEM
		if p.Namespace == "*" {
			return SCOPEALLPROJECT
		}
	case LEVELPROJECT:
		return fmt.Sprintf("/project/%d", projectID)
	}
}

func (p *Permission) fromScope(scope string, projectID int64) {
	if scope == SCOPESYSTEM {
		p.Kind = LEVELSYSTEM
		p.Namespace = "/"
	} else if scope == SCOPEALLPROJECT {
		p.Kind = LEVELPROJECT
		p.Namespace = "*"
	} else {
		p.Kind = LEVELPROJECT
		p.Namespace = "??"
	}
}

type Option struct {
	WithPermission bool
}
