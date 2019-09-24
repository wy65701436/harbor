package find

import "github.com/goharbor/harbor/src/pkg/art"

type ImmutableTagSelector interface {
	// Select ...
	Select(pid int64) ([]*art.Candidate, error)
}
