package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisCacheKey = "cache"

	hgetallAndDeleteScript = `
local items = redis.call('HGETALL', KEYS[1])
if #items > 0 then
	 local ok = pcall(redis.call, 'UNLINK', KEYS[1])
	 if not ok then redis.call('DEL', KEYS[1]) end
end
return items
`

	// getAndUpdateTTLScript atomically gets a hash field and updates its TTL.
	getAndUpdateTTLScript = `
local field = ARGV[1]
local deleteFlag = (ARGV[2] == "1" or ARGV[2] == "true")
local ttlTs = tonumber(ARGV[3]) or 0
local ttlDelta = tonumber(ARGV[4]) or 0

local value = redis.call('HGET', KEYS[1], field)
if not value then return false end

if deleteFlag then
  redis.call('HDEL', KEYS[1], field)
  return value
end

if ttlTs > 0 then
  redis.call('HExpireAt', KEYS[1], ttlTs, field)
elseif ttlDelta > 0 then
  local ttl = redis.call('HTTL', KEYS[1], field)
  if ttl < 0 then
    ttl = 0
  end
  local newTtl = ttl + ttlDelta
  redis.call('HExpire', KEYS[1], newTtl, field)
end

return value
`
)

// RedisConfig configures the Redis cache backend.
type RedisConfig struct {
	// Client is the Redis client to use.
	// If nil, a client is created from the URL.
	Client *redis.Client

	// URL is the Redis URL to use.
	// If empty, the Redis client is not created.
	URL string

	// Prefix is the prefix to use for all keys in the Redis cache.
	Prefix string

	// TTL is the time-to-live for all cache entries.
	TTL time.Duration
}

type RedisCache struct {
	client      *redis.Client
	ownedClient bool

	key string

	ttl time.Duration
}

func NewRedis(config RedisConfig) (*RedisCache, error) {
	if config.Prefix != "" && !strings.HasSuffix(config.Prefix, ":") {
		config.Prefix += ":"
	}

	if config.Client == nil && config.URL == "" {
		return nil, fmt.Errorf("%w: no redis client or url provided", ErrInvalidConfig)
	}

	client := config.Client
	if client == nil {
		opt, err := redis.ParseURL(config.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis url: %w", err)
		}

		client = redis.NewClient(opt)
	}

	return &RedisCache{
		client:      client,
		ownedClient: config.Client == nil,

		key: config.Prefix + redisCacheKey,

		ttl: config.TTL,
	}, nil
}

// Cleanup implements Cache.
func (r *RedisCache) Cleanup(_ context.Context) error {
	return nil
}

// Delete implements Cache.
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.HDel(ctx, r.key, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache item: %w", err)
	}

	return nil
}

// Drain implements Cache.
func (r *RedisCache) Drain(ctx context.Context) (map[string][]byte, error) {
	res, err := r.client.Eval(ctx, hgetallAndDeleteScript, []string{r.key}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to drain cache: %w", err)
	}

	arr, ok := res.([]any)
	if !ok || len(arr) == 0 {
		return map[string][]byte{}, nil
	}

	const itemsPerKey = 2
	out := make(map[string][]byte, len(arr)/itemsPerKey)
	for i := 0; i < len(arr); i += 2 {
		f, _ := arr[i].(string)
		v, _ := arr[i+1].(string)
		out[f] = []byte(v)
	}

	return out, nil
}

// Get implements Cache.
func (r *RedisCache) Get(ctx context.Context, key string, opts ...GetOption) ([]byte, error) {
	o := new(getOptions)
	o.apply(opts...)

	if o.isEmpty() {
		// No options, simple get
		val, err := r.client.HGet(ctx, r.key, key).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil, ErrKeyNotFound
			}

			return nil, fmt.Errorf("failed to get cache item: %w", err)
		}

		return []byte(val), nil
	}

	// Handle TTL options atomically using Lua script
	var ttlTimestamp, ttlDelta int64
	switch {
	case o.validUntil != nil:
		ttlTimestamp = o.validUntil.Unix()
	case o.setTTL != nil:
		ttlTimestamp = time.Now().Add(*o.setTTL).Unix()
	case o.updateTTL != nil:
		ttlDelta = int64(o.updateTTL.Seconds())
	case o.defaultTTL:
		ttlTimestamp = time.Now().Add(r.ttl).Unix()
	}

	delArg := "0"
	if o.delete {
		delArg = "1"
	}

	// Use atomic get and TTL update script
	result, err := r.client.Eval(ctx, getAndUpdateTTLScript, []string{r.key}, key, delArg, ttlTimestamp, ttlDelta).
		Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache item: %w", err)
	}

	if value, ok := result.(string); ok {
		return []byte(value), nil
	}

	return nil, ErrKeyNotFound
}

// GetAndDelete implements Cache.
func (r *RedisCache) GetAndDelete(ctx context.Context, key string) ([]byte, error) {
	return r.Get(ctx, key, AndDelete())
}

// Set implements Cache.
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, opts ...Option) error {
	options := new(options)
	if r.ttl > 0 {
		options.validUntil = time.Now().Add(r.ttl)
	}
	options.apply(opts...)

	_, err := r.client.Pipelined(ctx, func(p redis.Pipeliner) error {
		p.HSet(ctx, r.key, key, value)
		if !options.validUntil.IsZero() {
			p.HExpireAt(ctx, r.key, options.validUntil, key)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to set cache item: %w", err)
	}

	return nil
}

// SetOrFail implements Cache.
func (r *RedisCache) SetOrFail(ctx context.Context, key string, value []byte, opts ...Option) error {
	val, err := r.client.HSetNX(ctx, r.key, key, value).Result()
	if err != nil {
		return fmt.Errorf("failed to set cache item: %w", err)
	}

	if !val {
		return ErrKeyExists
	}

	options := new(options)
	if r.ttl > 0 {
		options.validUntil = time.Now().Add(r.ttl)
	}
	options.apply(opts...)

	if !options.validUntil.IsZero() {
		if expErr := r.client.HExpireAt(ctx, r.key, options.validUntil, key).Err(); expErr != nil {
			return fmt.Errorf("failed to set cache item ttl: %w", expErr)
		}
	}

	return nil
}

func (r *RedisCache) Close() error {
	if r.ownedClient {
		if err := r.client.Close(); err != nil {
			return fmt.Errorf("failed to close redis client: %w", err)
		}
	}

	return nil
}

var _ Cache = (*RedisCache)(nil)
