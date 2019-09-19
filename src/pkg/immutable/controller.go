package immutable

import (
	"errors"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
)

var ErrTagUnknown = errors.New("unknown tag")

// APIController to handle the requests related with immutable
type APIController interface {
	GetImmutableRule(id int64) (*policy.Metadata, error)

	CreateRetentionImmutableRule(p *policy.Metadata) (int64, error)

	UpdateImmutableRule(p *policy.Metadata) error

	DeleteImmutableRule(id int64) error

	Match()
}

// DefaultAPIController ...
type DefaultAPIController struct {
	projectManager project.Manager
	repositoryMgr  repository.Manager
}
