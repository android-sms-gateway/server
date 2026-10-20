package cache

import (
	"sync"
	"time"

	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/cachefx/cache"
	"go.uber.org/zap"
)

// NewTestFactory exposes the package-private module wiring to the external test
// package, backed by the in-memory cachefx implementation.
func NewTestFactory(name string) (Factory, error) {
	inner, err := cachefx.NewFactory(cachefx.Config{URL: "memory://"})
	if err != nil {
		return nil, err
	}

	return NewTestFactoryWith(inner.WithName(name)), nil
}

// NewTestFactoryWith exposes the package-private tracking factory to the
// external test package, so a test can substitute its own cache implementation.
func NewTestFactoryWith(inner cachefx.Factory) Factory {
	return &trackingFactory{
		Factory: inner,

		tracker: &cacheTracker{
			mu:     sync.Mutex{},
			caches: []cache.Cache{},
		},
	}
}

// NewTestSweeper exposes the package-private sweeper constructor to the
// external test package with a configurable interval.
func NewTestSweeper(factory Factory, interval time.Duration) *Sweeper {
	return &Sweeper{factory: factory, interval: interval, logger: zap.NewNop()}
}
