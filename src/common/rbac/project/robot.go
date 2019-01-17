package project

import "github.com/goharbor/harbor/src/common/rbac"

// robotContext the context interface for the robot
type robotContext interface {
	// Index whether the robot is authenticated
	IsAuthenticated() bool
	// GetUsername returns the name of robot
	GetUsername() string
	// GetPolicy get the rbac policies from security context
	GetPolicies() []*rbac.Policy
}

// robot implement the rbac.User interface for project robot account
type robot struct {
	ctx       robotContext
	namespace rbac.Namespace
}

/// GetUserName get the robot name.
func (r *robot) GetUserName() string {
	return r.ctx.GetUsername()
}

// GetPolicies ...
func (r *robot) GetPolicies() []*rbac.Policy {
	if r.namespace.IsPublic() {
		return policiesForPublicProject(r.namespace)
	}
	return r.ctx.GetPolicies()
}

// GetRoles robot has no definition of role, always return nil here.
func (r *robot) GetRoles() []rbac.Role {
	return nil
}

// NewRobot ...
func NewRobot(ctx robotContext, namespace rbac.Namespace) rbac.User {
	return &robot{
		ctx:       ctx,
		namespace: namespace,
	}
}
