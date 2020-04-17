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
	"github.com/goharbor/harbor/src/lib/log"
	"sync"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/admin"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	user      *models.User
	pm        promgr.ProjectManager
	evaluator evaluator.Evaluator
	once      sync.Once
}

// NewSecurityContext ...
func NewSecurityContext(user *models.User, pm promgr.ProjectManager) *SecurityContext {
	return &SecurityContext{
		user: user,
		pm:   pm,
	}
}

// Name returns the name of the security context
func (s *SecurityContext) Name() string {
	return "local"
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
	return s.user.Username
}

// User get the current user
func (s *SecurityContext) User() *models.User {
	return s.user
}

// IsSysAdmin returns whether the authenticated user is system admin
// It returns false if the user has not been authenticated
func (s *SecurityContext) IsSysAdmin() bool {
	log.Info("IsSysAdmin .......")
	if !s.IsAuthenticated() {
		return false
	}
	log.Info("IsSysAdmin user %v.......", s.user)
	log.Info("IsSysAdmin user1 %v.......", s.user.SysAdminFlag)
	log.Info("IsSysAdmin user2 %v.......", s.user.AdminRoleInAuth)
	log.Info("IsSysAdmin user3 %v.......", s.user.HasAdminRole)
	return s.user.SysAdminFlag || s.user.AdminRoleInAuth
}

// IsSolutionUser ...
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// Can returns whether the user can do action on resource
func (s *SecurityContext) Can(action types.Action, resource types.Resource) bool {
	s.once.Do(func() {
		var evaluators evaluator.Evaluators
		if s.IsSysAdmin() {
			evaluators = evaluators.Add(admin.New(s.GetUsername()))
		}
		evaluators = evaluators.Add(rbac.NewProjectUserEvaluator(s.User(), s.pm))

		s.evaluator = evaluators
	})

	return s.evaluator != nil && s.evaluator.HasPermission(resource, action)
}
