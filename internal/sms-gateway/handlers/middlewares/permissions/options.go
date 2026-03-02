package permissions

type options struct {
	exact bool
}

type Option func(*options)

func WithExact() Option {
	return func(o *options) {
		o.exact = true
	}
}
