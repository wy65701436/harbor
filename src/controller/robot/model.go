package robot

import (
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

type Robot struct {
	model.Robot
	ProjectName string
	Level       string
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	Kind      string          `json:"kind"`
	Namespace string          `json:"namespace"`
	Access    []*types.Policy `json:"access"`
}

type Option struct {
	WithPermission bool
}
