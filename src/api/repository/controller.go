// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repository

import (
	"context"
	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
)

var (
	// Ctl is a global repository controller instance
	Ctl = NewController()
)

// Controller defines the operations related with repositories
type Controller interface {
	// Ensure the repository specified by the "name" exists, creates it if it doesn't exist.
	// The "name" should contain the namespace part. The "created" will be set as true
	// when the repository is created
	Ensure(ctx context.Context, name string) (created bool, id int64, err error)
	// Count returns the total count of repositories according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List repositories according to the query
	List(ctx context.Context, query *q.Query) (repositories []*models.RepoRecord, err error)
	// Get the repository specified by ID
	Get(ctx context.Context, id int64) (repository *models.RepoRecord, err error)
	// GetByName gets the repository specified by name
	GetByName(ctx context.Context, name string) (repository *models.RepoRecord, err error)
	// Delete the repository specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Update the repository. Specify the properties or all properties will be updated
	Update(ctx context.Context, repository *models.RepoRecord, properties ...string) (err error)
}

// NewController creates an instance of the default repository controller
func NewController() Controller {
	return &controller{
		proMgr:  project.Mgr,
		repoMgr: repository.Mgr,
		artCtl:  artifact.Ctl,
	}
}

type controller struct {
	proMgr  project.Manager
	repoMgr repository.Manager
	artCtl  artifact.Controller
}

func (c *controller) Ensure(ctx context.Context, name string) (bool, int64, error) {
	query := &q.Query{
		Keywords: map[string]interface{}{
			"name": name,
		},
	}
	repositories, err := c.repoMgr.List(ctx, query)
	if err != nil {
		return false, 0, err
	}
	// the repository already exists, return directly
	if len(repositories) > 0 {
		return false, repositories[0].RepositoryID, nil
	}

	// the repository doesn't exist, create it first
	projectName, _ := utils.ParseRepository(name)
	project, err := c.proMgr.Get(projectName)
	if err != nil {
		return false, 0, err
	}
	id, err := c.repoMgr.Create(ctx, &models.RepoRecord{
		ProjectID: project.ProjectID,
		Name:      name,
	})
	if err != nil {
		// if got conflict error, try to get again
		if ierror.IsConflictErr(err) {
			repositories, err = c.repoMgr.List(ctx, query)
			if err != nil {
				return false, 0, err
			}
			if len(repositories) > 0 {
				return false, repositories[0].RepositoryID, nil
			}
		}
		return false, 0, err
	}
	return true, id, nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.repoMgr.Count(ctx, query)
}

func (c *controller) List(ctx context.Context, query *q.Query) ([]*models.RepoRecord, error) {
	return c.repoMgr.List(ctx, query)
}

func (c *controller) Get(ctx context.Context, id int64) (*models.RepoRecord, error) {
	return c.repoMgr.Get(ctx, id)
}

func (c *controller) GetByName(ctx context.Context, name string) (*models.RepoRecord, error) {
	return c.repoMgr.GetByName(ctx, name)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	// TODO auth
	// TODO how to make sure the logic included by middlewares(immutable, readonly, quota, etc)
	// TODO is covered when deleting the artifacts of the repository
	artifacts, err := c.artCtl.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": id,
		},
	}, nil)
	if err != nil {
		return err
	}
	for _, artifact := range artifacts {
		if err = c.artCtl.Delete(ctx, artifact.ID); err != nil {
			return err
		}
	}
	return c.repoMgr.Delete(ctx, id)

	// TODO fire event
}

func (c *controller) Update(ctx context.Context, repository *models.RepoRecord, properties ...string) error {
	return c.repoMgr.Update(ctx, repository, properties...)
}
