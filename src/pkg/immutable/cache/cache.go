package cache

import "github.com/goharbor/harbor/src/pkg/art"

type Cache interface {
	Set(pid int64, c art.Candidate) error

	Stat(pid int64, digest string) (bool, error)

	Clear(pid int64, c art.Candidate) error

	Flush(pid int64) error
}
