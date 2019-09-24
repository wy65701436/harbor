package rule

import (
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/art/selectors/index"
	"github.com/goharbor/harbor/src/pkg/immutable"
	"github.com/goharbor/harbor/src/pkg/immutable/rule"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
)

type RuleSelector struct {
	projectMgr    project.Manager
	repositoryMgr repository.Manager
	rules         []rule.IMRule
}

func NewRuleSelector() RuleSelector {
	return RuleSelector{}
}

func (rh *RuleSelector) Select(pid int64) ([]*art.Candidate, error) {
	var repositoryCandidates []*art.Candidate

	for _, rule := range rh.rules {
		if rule.Disabled {
			continue
		}

		repositories, err := getRepositories(rh.projectMgr, rh.repositoryMgr, pid)
		if err != nil {
			return repositoryCandidates, err
		}

		for _, repository := range repositories {
			repositoryCandidates = append(repositoryCandidates, repository)
		}

		// filter repositories according to the repository selectors
		for _, repositorySelector := range rule.Metadata.RepoSelectors["repository"] {
			selector, err := index.Get(repositorySelector.Kind, repositorySelector.Decoration,
				repositorySelector.Pattern)
			if err != nil {
				return repositoryCandidates, err
			}
			repositoryCandidates, err = selector.Select(repositoryCandidates)
			if err != nil {
				return repositoryCandidates, err
			}
		}

	}

	return repositoryCandidates, nil
}

func (rh *RuleSelector) getImmutableRules(pid int64) error {
	rules, err := immutable.NewDefaultRuleManager().QueryEnabledImmutableRuleByProjectID(pid)
	if err != nil {
		return err
	}
	rh.rules = rules
	return nil
}

// only image is supported
func getRepositories(projectMgr project.Manager, repositoryMgr repository.Manager,
	projectID int64) ([]*art.Candidate, error) {
	var candidates []*art.Candidate
	// get image repositories
	imageRepositories, err := repositoryMgr.ListImageRepositories(projectID)
	if err != nil {
		return nil, err
	}
	for _, r := range imageRepositories {
		namespace, repo := utils.ParseRepository(r.Name)
		candidates = append(candidates, &art.Candidate{
			Namespace:  namespace,
			Repository: repo,
			Kind:       "image",
		})
	}
	return candidates, nil
}
