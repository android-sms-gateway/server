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
// It matches the shortest TTL handed out by the application (the one-time codes
// and the queue limiter use 60s), so an expired entry is reclaimed within one
// interval after it expires.
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

// cacheTracker holds the caches handed out by a tracking factory.
//
// It is shared between a factory and the factories derived from it, so a cache
// is swept no matter which of them created it.
type cacheTracker struct {
	mu     sync.Mutex
	caches []cache.Cache
}

func (t *cacheTracker) add(c cache.Cache) {
	t.mu.Lock()
	t.caches = append(t.caches, c)
	t.mu.Unlock()
}

func (t *cacheTracker) list() []cache.Cache {
	t.mu.Lock()
	defer t.mu.Unlock()

	caches := make([]cache.Cache, len(t.caches))
	copy(caches, t.caches)

	return caches
}

type trackingFactory struct {
	cachefx.Factory

	tracker *cacheTracker
}

func (f *trackingFactory) New(name string) (cache.Cache, error) {
	inner, err := f.Factory.New(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache %q: %w", name, err)
	}

	f.tracker.add(inner)

	return inner, nil
}

// WithName returns a factory that shares the tracker with its parent, so the
// caches it hands out are swept by the parent's CleanupAll as well.
func (f *trackingFactory) WithName(prefix string) cachefx.Factory {
	return &trackingFactory{
		Factory: f.Factory.WithName(prefix),

		tracker: f.tracker,
	}
}

func (f *trackingFactory) CleanupAll(ctx context.Context) error {
	var errs []error
	for _, c := range f.tracker.list() {
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
