package api

import (
	"github.com/goharbor/harbor/src/pkg/immutable/rule"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
)

// APIController to handle the requests related with immutable
type APIController interface {
	// GetImmutableRule ...
	ListImmutableRules(id int64) (*[]rule.Metadata, error)

	// CreateImmutableRule ...
	CreateImmutableRule(p *rule.Metadata) (int64, error)

	// UpdateImmutableRule ...
	UpdateImmutableRule(p *rule.Metadata) error

	// DeleteImmutableRule ...
	DeleteImmutableRule(id int64) error
}

// DefaultAPIController ...
type DefaultAPIController struct {
	projectManager project.Manager
	repositoryMgr  repository.Manager
}

// NewAPIController ...
func NewAPIController(projectManager project.Manager, repositoryMgr repository.Manager) APIController {
	return &DefaultAPIController{
		projectManager: projectManager,
		repositoryMgr:  repositoryMgr,
	}
}

// GetImmutableRule ...
func (d *DefaultAPIController) ListImmutableRules(id int64) (*[]rule.Metadata, error) {
	return nil, nil
}

// CreateImmutableRule ...
func (d *DefaultAPIController) CreateImmutableRule(p *rule.Metadata) (int64, error) {
	return 0, nil
}

// UpdateImmutableRule ...
func (d *DefaultAPIController) UpdateImmutableRule(p *rule.Metadata) error {
	return nil
}

// DeleteImmutableRule ...
func (d *DefaultAPIController) DeleteImmutableRule(id int64) error {
	return nil
}
