package find

type Hitter interface {
	Hit(pid int64, repository string, tag string) (bool, error)
}
