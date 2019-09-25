package rule

import (
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/art/selectors/index"
	"github.com/goharbor/harbor/src/pkg/immutable"
	"github.com/goharbor/harbor/src/pkg/immutable/match"
	"github.com/goharbor/harbor/src/pkg/immutable/rule"

	"encoding/json"
)

// Matcher ...
type Matcher struct {
	pid   int64
	rules []rule.Metadata
}

// Match ...
func (rm *Matcher) Match(uploads []*art.Candidate) (bool, error) {
	for _, r := range rm.rules {
		if r.Disabled {
			continue
		}

		// match repositories according to the repository selectors
		var repositoryCandidates []*art.Candidate
		for _, repositorySelector := range r.ScopeSelectors["repository"] {
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

func (rm *Matcher) getImmutableRules() error {
	rules, err := immutable.NewDefaultRuleManager().QueryEnabledImmutableRuleByProjectID(rm.pid)
	if err != nil {
		return err
	}
	for _, r := range rules {
		rmeta := rule.Metadata{}
		if err := json.Unmarshal([]byte(r.TagFilter), &rmeta); err != nil {
			return err
		}
		rm.rules = append(rm.rules, rmeta)
	}
	return nil
}

// NewRuleMatcher ...
func NewRuleMatcher(pid int64) matcher.ImmutableTagMatcher {
	return &Matcher{
		pid: pid,
	}
}
