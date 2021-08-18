package user

type UserOption func(*Options)

type Options struct {
	IncludeAdmin bool
}

// WithAdmin set the IncludeAdmin = true
func WithAdmin() UserOption {
	return func(o *Options) {
		o.IncludeAdmin = true
	}
}

func newOptions(options ...UserOption) *Options {
	opts := &Options{}
	for _, f := range options {
		f(opts)
	}
	return opts
}
