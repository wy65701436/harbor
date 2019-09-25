package rule

import (
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/art/selectors/index"
	"github.com/goharbor/harbor/src/pkg/immutable"
	"github.com/goharbor/harbor/src/pkg/immutable/rule"
)

// RuleMatcher ...
type RuleMatcher struct {
	pid   int64
	rules []rule.Metadata
}

// Match ...
func (rm *RuleMatcher) Match(uploads []*art.Candidate) (bool, error) {
	for _, r := range rm.rules {
		if r.Disabled {
			continue
		}

		// match repositories according to the repository selectors
		var repositoryCandidates []*art.Candidate
		for _, repositorySelector := range r.RepoSelectors["repository"] {
			selector, err := index.Get(repositorySelector.Kind, repositorySelector.Decoration,
				repositorySelector.Pattern)
			if err != nil {
				return false, err
			}
			repositoryCandidates, err = selector.Select(uploads)
			if err != nil {
				return false, err
			}
		}

		if len(repositoryCandidates) == 0 {
			continue
		}

		// match tag according to the tag selectors
		var tagCandidates []*art.Candidate
		for _, tagSelector := range r.TagSelectors {
			selector, err := index.Get(tagSelector.Kind, tagSelector.Decoration,
				tagSelector.Pattern)
			if err != nil {
				return false, err
			}
			tagCandidates, err = selector.Select(uploads)
			if err != nil {
				return false, err
			}
		}

		if len(tagCandidates) == 0 {
			continue
		}
		return true, nil
	}
	return false, nil
}

func (rm *RuleMatcher) getImmutableRules() error {
	rules, err := immutable.NewDefaultRuleManager().QueryEnabledImmutableRuleByProjectID(rm.pid)
	if err != nil {
		return err
	}
	rm.rules = rules
	return nil
}

// NewRuleMatcher ...
func NewRuleMatcher(pid int64) RuleMatcher {
	return RuleMatcher{
		pid: pid,
	}
}
