package rule

import (
	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/pkg/immutable/dao/model"
)

// Immutable rule
type IMRule struct {
	model.ImmutableRule

	// Metadata of the immutable rule
	Metadata Metadata
}

// Metadata of the immutable rule
type Metadata struct {
	// UUID of rule
	ID int `json:"id"`

	// Disabled rule
	Disabled bool `json:"disabled"`

	// TagSelectors attached to the rule for filtering tags
	TagSelectors []*Selector `json:"tag_selectors" valid:"Required"`

	// RepoSelectors attached to the rule for filtering scope (e.g: repositories or namespaces)
	RepoSelectors map[string][]*Selector `json:"repo_selectors" valid:"Required"`
}

// Valid Valid
func (m *Metadata) Valid(v *validation.Validation) {
	for _, ts := range m.TagSelectors {
		if pass, _ := v.Valid(ts); !pass {
			return
		}
	}
	for _, ss := range m.RepoSelectors {
		for _, s := range ss {
			if pass, _ := v.Valid(s); !pass {
				return
			}
		}
	}
}

// Selector to narrow down the list
type Selector struct {
	// Kind of the selector
	// "doublestar" or "label"
	Kind string `json:"kind" valid:"Required;Match(doublestar)"`

	// Decorated the selector
	// for "doublestar" : "matching" and "excluding"
	// for "label" : "with" and "without"
	Decoration string `json:"decoration" valid:"Required"`

	// Param for the selector
	Pattern string `json:"pattern" valid:"Required"`
}

// Parameters of rule, indexed by the key
type Parameters map[string]Parameter

// Parameter of rule
type Parameter interface{}
