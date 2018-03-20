package middlewares

import (
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/clair"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/promgr"
)

// policyChecker checks the policy of a project by project name, to determine if it's needed to check the image's status under this project.
type policyChecker interface {
	// contentTrustEnabled returns whether a project has enabled content trust.
	contentTrustEnabled(name string) bool
	// vulnerablePolicy  returns whether a project has enabled vulnerable, and the project's severity.
	vulnerablePolicy(name string) (bool, models.Severity)
}

type pmsPolicyChecker struct {
	pm promgr.ProjectManager
}

func (pc pmsPolicyChecker) contentTrustEnabled(name string) bool {
	project, err := pc.pm.Get(name)
	if err != nil {
		log.Errorf("Unexpected error when getting the project, error: %v", err)
		return true
	}
	return project.ContentTrustEnabled()
}
func (pc pmsPolicyChecker) vulnerablePolicy(name string) (bool, models.Severity) {
	project, err := pc.pm.Get(name)
	if err != nil {
		log.Errorf("Unexpected error when getting the project, error: %v", err)
		return true, models.SevUnknown
	}
	return project.VulPrevented(), clair.ParseClairSev(project.Severity())
}

// newPMSPolicyChecker returns an instance of an pmsPolicyChecker
func newPMSPolicyChecker(pm promgr.ProjectManager) policyChecker {
	return &pmsPolicyChecker{
		pm: pm,
	}
}

func GetPolicyChecker() policyChecker {
	return newPMSPolicyChecker(config.GlobalProjectMgr)
}
