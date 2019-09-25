package matcher

import "github.com/goharbor/harbor/src/pkg/art"

type ImmutableTagMatcher interface {
	// Match whether the candidate is in the immutable list
	Match([]*art.Candidate) (bool, error)
}
