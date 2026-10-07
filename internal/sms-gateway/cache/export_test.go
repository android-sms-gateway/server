package cache

import (
	"time"

	"github.com/go-core-fx/cachefx"
	"go.uber.org/zap"
)

// NewTestFactory exposes the package-private module wiring to the external test
// package, backed by the in-memory cachefx implementation.
func NewTestFactory(name string) (Factory, error) {
	inner, err := cachefx.NewFactory(cachefx.Config{URL: "memory://"})
	if err != nil {
		return nil, err
	}

	return &trackingFactory{Factory: inner.WithName(name)}, nil
}

// NewTestSweeper exposes the package-private sweeper constructor to the
// external test package with a configurable interval.
func NewTestSweeper(factory Factory, interval time.Duration) *Sweeper {
	return &Sweeper{factory: factory, interval: interval, logger: zap.NewNop()}
}
