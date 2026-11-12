package messages

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	cacheImpl "github.com/go-core-fx/cachefx/cache"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// countCacheKeyVersion namespaces cached list totals. Bump it whenever the
// counted result changes meaning, so entries written by an older release are
// never read back.
const countCacheKeyVersion = "count:v1:"

// countCacheKey namespaces a filter hash into the key a total is stored under.
func countCacheKey(hash string) string {
	return countCacheKeyVersion + hash
}

type countCache struct {
	ttl     time.Duration
	storage cacheImpl.Cache
	group   singleflight.Group
	logger  *zap.Logger
}

func newCountCache(config Config, storage cacheImpl.Cache, logger *zap.Logger) *countCache {
	return &countCache{
		ttl:     config.ListCountCacheTTL,
		storage: storage,
		group:   singleflight.Group{},
		logger:  logger,
	}
}

// Count returns the list total stored for hash. A hit is reported only when
// this call read the value from the cache; concurrent misses share a single
// fetch instead of repeating it for every caller. A non-positive TTL disables
// storage while keeping the coalescing, so no total can outlive its freshness
// window.
func (c *countCache) Count(hash string, fetch func() (int64, error)) (int64, bool, error) {
	key := countCacheKey(hash)

	if total, ok := c.readLogged(key); ok {
		return total, true, nil
	}

	value, err, _ := c.group.Do(key, func() (any, error) {
		total, loadErr := c.load(key, fetch)
		if loadErr != nil {
			return nil, loadErr
		}

		return total, nil
	})
	if err != nil {
		return 0, false, fmt.Errorf("failed to fetch list total: %w", err)
	}

	total, ok := value.(int64)
	if !ok {
		return 0, false, ErrCountCacheResult
	}

	return total, false, nil
}

// load fills a single coalesced flight: it re-reads the cache, so a request
// that waited for another one never repeats its database work.
func (c *countCache) load(key string, fetch func() (int64, error)) (int64, error) {
	if total, ok := c.readLogged(key); ok {
		return total, nil
	}

	total, err := fetch()
	if err != nil {
		return 0, err
	}

	if setErr := c.set(key, total); setErr != nil {
		c.logger.Warn("failed to cache list total", zap.String("key", key), zap.Error(setErr))
	}

	return total, nil
}

// readLogged reads the total and reports any cache failure as a miss, so a
// broken cache degrades to the database instead of failing the request.
func (c *countCache) readLogged(key string) (int64, bool) {
	total, ok, err := c.read(key)
	if err != nil {
		c.logger.Warn("failed to read list total from cache", zap.String("key", key), zap.Error(err))
		return 0, false
	}

	return total, ok
}

func (c *countCache) read(key string) (int64, bool, error) {
	if c.ttl <= 0 {
		return 0, false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cacheTimeout)
	defer cancel()

	// The TTL is intentionally not renewed on read: a continuously polled total
	// would otherwise never be refreshed and would pin a stale value forever.
	data, err := c.storage.Get(ctx, key)
	if err != nil {
		if errors.Is(err, cacheImpl.ErrKeyNotFound) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("failed to get list total from cache: %w", err)
	}

	if len(data) == 0 {
		return 0, false, nil
	}

	total, parseErr := strconv.ParseInt(string(data), 10, 64)
	if parseErr != nil {
		return 0, false, fmt.Errorf("failed to parse cached list total: %w", parseErr)
	}

	return total, true, nil
}

func (c *countCache) set(key string, total int64) error {
	if c.ttl <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cacheTimeout)
	defer cancel()

	if err := c.storage.Set(
		ctx,
		key,
		[]byte(strconv.FormatInt(total, 10)),
		cacheImpl.WithTTL(c.ttl),
	); err != nil {
		return fmt.Errorf("failed to set list total in cache: %w", err)
	}

	return nil
}
