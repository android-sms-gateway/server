package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/cachefx/cache"
	"go.uber.org/zap"
)

// sweeperInterval is how often the sweeper evicts expired entries from every
// cache created by the Factory.
//
// The interval must be shorter than the shortest TTL handed out by the
// application (the limiter uses 60s), otherwise expired entries accumulate
// between sweeps and the process keeps the memory until the next tick.
const sweeperInterval = time.Minute

// Factory is the cachefx factory, extended with a group sweep for every cache
// it has handed out.
//
// Eviction of expired entries is not guaranteed by every backend: the in-memory
// backend keeps expired items until an explicit Cleanup, and each New call
// returns a distinct instance, so a sweeper built on its own instances would
// sweep maps nobody reads.
type Factory interface {
	cachefx.Factory

	// CleanupAll evicts expired entries from every cache handed out by this
	// factory.
	CleanupAll(ctx context.Context) error
}

type trackingFactory struct {
	cachefx.Factory

	cachesMu sync.Mutex
	caches   []cache.Cache
}

func (f *trackingFactory) New(name string) (cache.Cache, error) {
	inner, err := f.Factory.New(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache %q: %w", name, err)
	}

	f.cachesMu.Lock()
	f.caches = append(f.caches, inner)
	f.cachesMu.Unlock()

	return inner, nil
}

func (f *trackingFactory) WithName(prefix string) cachefx.Factory {
	return &trackingFactory{
		Factory: f.Factory.WithName(prefix),

		cachesMu: sync.Mutex{},
		caches:   []cache.Cache{},
	}
}

func (f *trackingFactory) CleanupAll(ctx context.Context) error {
	f.cachesMu.Lock()
	caches := make([]cache.Cache, len(f.caches))
	copy(caches, f.caches)
	f.cachesMu.Unlock()

	var errs []error
	for _, c := range caches {
		if err := c.Cleanup(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cache cleanup failed: %w", errs[0])
	}

	return nil
}

// Sweeper periodically evicts expired entries from every cache created by the
// Factory.
type Sweeper struct {
	factory  Factory
	interval time.Duration
	logger   *zap.Logger
}

func NewSweeper(factory Factory, logger *zap.Logger) *Sweeper {
	return &Sweeper{
		factory:  factory,
		interval: sweeperInterval,
		logger:   logger,
	}
}

func (s *Sweeper) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Cache sweeper stopped")
			return nil
		case <-ticker.C:
			if err := s.factory.CleanupAll(ctx); err != nil {
				s.logger.Error("failed to cleanup caches", zap.Error(err))
			}
		}
	}
}
