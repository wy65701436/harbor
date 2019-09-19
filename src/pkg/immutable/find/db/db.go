package db

import (
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/art/selectors/index"
	"github.com/goharbor/harbor/src/pkg/immutable/rule"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
)

type DBHitter struct {
	pid        int64
	repository string
	tag        string

	projectMgr    project.Manager
	repositoryMgr repository.Manager
	rules         []rule.IMRule
}

func NewDBHitter(pid int64, repository string, tag string) DBHitter {
	return DBHitter{
		pid:        pid,
		repository: repository,
		tag:        tag,
	}
}

func (db *DBHitter) Hit() (bool, error) {
	var repositoryCandidates []*art.Candidate

	for _, rule := range db.rules {
		if rule.Disabled {
			continue
		}

		repositories, err := getRepositories(db.projectMgr, db.repositoryMgr, pid)
		if err != nil {
			return false, err
		}

		for _, repository := range repositories {
			repositoryCandidates = append(repositoryCandidates, repository)
		}

		// filter repositories according to the repository selectors
		for _, repositorySelector := range rule.Metadata.ScopeSelectors["repository"] {
			selector, err := index.Get(repositorySelector.Kind, repositorySelector.Decoration,
				repositorySelector.Pattern)
			if err != nil {
				return false, err
			}
			repositoryCandidates, err = selector.Select(repositoryCandidates)
			if err != nil {
				return false, err
			}
		}

	}

	if len(repositoryCandidates) == 0 {
		return false, nil
	}

	for _, c := range repositoryCandidates {
		if c.Repository == db.repository && c.Tag == db.tag {
			return true, nil
		}
	}

	return false, nil
}

func (db *DBHitter) getImmutableRules() error {
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
