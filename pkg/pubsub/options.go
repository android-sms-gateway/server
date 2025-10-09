package pubsub

type Option func(*options)

type options struct {
	bufferSize uint
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}

	return o
}

func WithBufferSize(bufferSize uint) Option {
	return func(o *options) {
		o.bufferSize = bufferSize
	}
}
