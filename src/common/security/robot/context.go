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

package local

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	user *models.Robot
}

// NewSecurityContext ...
func NewSecurityContext(user *models.Robot) *SecurityContext {
	return &SecurityContext{
		user: user,
	}
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	return s.user != nil
}

// GetUsername returns the username of the authenticated user
// It returns null if the user has not been authenticated
func (s *SecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	return s.user.Name
}

// IsSysAdmin robot cannot be a system admin
func (s *SecurityContext) IsSysAdmin() bool {
	return false
}

// IsSolutionUser ...
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// HasReadPerm returns whether the user has read permission to the project
func (s *SecurityContext) HasReadPerm(projectIDOrName interface{}) bool {
	isPublicProject, _ := s.pm.IsPublic(projectIDOrName)
	return s.Can(project.ActionPull, rbac.NewProjectNamespace(projectIDOrName, isPublicProject).Resource(project.ResourceImage))
}

// HasWritePerm returns whether the user has write permission to the project
func (s *SecurityContext) HasWritePerm(projectIDOrName interface{}) bool {
	return false
}

// HasAllPerm returns whether the user has all permissions to the project
func (s *SecurityContext) HasAllPerm(projectIDOrName interface{}) bool {
	return false
}

// GetMyProjects ...
func (s *SecurityContext) GetMyProjects() ([]*models.Project, error){
	return nil, nil
}

// GetProjectRoles ...
func (s *SecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	return nil
}

// Can returns whether the user can do action on resource
func (s *SecurityContext) Can(action rbac.Action, resource rbac.Resource) bool {
	ns, err := resource.GetNamespace()
	if err == nil {
		switch ns.Kind() {
		case "project":
			projectIDOrName := ns.Identity()
			isPublicProject, _ := s.pm.IsPublic(projectIDOrName)
			projectNamespace := rbac.NewProjectNamespace(projectIDOrName, isPublicProject)
			user := project.NewUser(s, projectNamespace, s.GetProjectRoles(projectIDOrName)...)
			return rbac.HasPermission(user, resource, action)
		}
	}

	return false
}