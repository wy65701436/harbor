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

package robot

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestIsAuthenticated(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil, nil)
	assert.False(t, ctx.IsAuthenticated())

	// authenticated
	ctx = NewSecurityContext(&models.Robot{
		Name:     "test",
		Disabled: false,
	}, nil, nil)
	assert.True(t, ctx.IsAuthenticated())
}

func TestGetUsername(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil, nil)
	assert.Equal(t, "", ctx.GetUsername())

	// authenticated
	ctx = NewSecurityContext(&models.Robot{
		Name:     "test",
		Disabled: false,
	}, nil, nil)
	assert.Equal(t, "test", ctx.GetUsername())
}

func TestIsSysAdmin(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil, nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, non admin
	ctx = NewSecurityContext(&models.Robot{
		Name:     "test",
		Disabled: false,
	}, nil, nil)
	assert.False(t, ctx.IsSysAdmin())
}

func TestIsSolutionUser(t *testing.T) {
	ctx := NewSecurityContext(nil, nil, nil)
	assert.False(t, ctx.IsSolutionUser())
}

func TestHasReadPerm(t *testing.T) {
	// public project
	rbacPolicy := &rbac.Policy{
		Resource: "/project/library/image",
		Action:   "pull",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	ctx := NewSecurityContext(nil, pm, policies)
	assert.True(t, ctx.HasReadPerm("library"))
}

func TestHasWritePerm(t *testing.T) {
	// unauthenticated
	rbacPolicy := &rbac.Policy{
		Resource: "/project/library/image",
		Action:   "push",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	ctx := NewSecurityContext(nil, pm, policies)
	assert.True(t, ctx.HasWritePerm("library"))
}

func TestHasAllPerm(t *testing.T) {
	// unauthenticated
	rbacPolicy := &rbac.Policy{
		Resource: "/project/library/image",
		Action:   "pull+push",
	}
	policies := []*rbac.Policy{}
	policies = append(policies, rbacPolicy)

	ctx := NewSecurityContext(nil, pm, policies)
	assert.True(t, ctx.HasAllPerm("library"))
}

func TestGetMyProjects(t *testing.T) {
	ctx := NewSecurityContext(nil, nil, nil)
	projects, err := ctx.GetMyProjects()
	require.Nil(t, err)
	assert.Nil(t, projects)
}

func TestGetProjectRoles(t *testing.T) {
	ctx := NewSecurityContext(nil, nil, nil)
	roles := ctx.GetProjectRoles("test")
	assert.Nil(t, roles)
}
