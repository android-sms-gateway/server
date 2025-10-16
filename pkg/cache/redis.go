package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisCacheKey = "cache"

	// getAndDeleteScript atomically gets and deletes a hash field
	getAndDeleteScript = `
local value = redis.call('HGET', KEYS[1], ARGV[1])
if value then
	redis.call('HDEL', KEYS[1], ARGV[1])
	return value
else
	return false
end
`

	hgetallAndDeleteScript = `
local items = redis.call('HGETALL', KEYS[1])
if #items > 0 then
  local ok = pcall(redis.call, 'UNLINK', KEYS[1])
  if not ok then redis.call('DEL', KEYS[1]) end
end
return items
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

type redisCache struct {
	client      *redis.Client
	ownedClient bool

	key string

	ttl time.Duration
}

func NewRedis(config RedisConfig) (*redisCache, error) {
	if config.Prefix != "" && !strings.HasSuffix(config.Prefix, ":") {
		config.Prefix += ":"
	}

	if config.Client == nil && config.URL == "" {
		return nil, fmt.Errorf("no redis client or url provided")
	}

	client := config.Client
	if client == nil {
		opt, err := redis.ParseURL(config.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis url: %w", err)
		}

		client = redis.NewClient(opt)
	}

	return &redisCache{
		client:      client,
		ownedClient: config.Client == nil,

		key: config.Prefix + redisCacheKey,

		ttl: config.TTL,
	}, nil
}

// Cleanup implements Cache.
func (r *redisCache) Cleanup(_ context.Context) error {
	return nil
}

// Delete implements Cache.
func (r *redisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.HDel(ctx, r.key, key).Err(); err != nil {
		return fmt.Errorf("can't delete cache item: %w", err)
	}

	return nil
}

// Drain implements Cache.
func (r *redisCache) Drain(ctx context.Context) (map[string]string, error) {
	res, err := r.client.Eval(ctx, hgetallAndDeleteScript, []string{r.key}).Result()
	if err != nil {
		return nil, fmt.Errorf("can't drain cache: %w", err)
	}

	arr, ok := res.([]any)
	if !ok || len(arr) == 0 {
		return map[string]string{}, nil
	}

	out := make(map[string]string, len(arr)/2)
	for i := 0; i < len(arr); i += 2 {
		f, _ := arr[i].(string)
		v, _ := arr[i+1].(string)
		out[f] = v
	}

	return out, nil
}

// Get implements Cache.
func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.HGet(ctx, r.key, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}

		return "", fmt.Errorf("can't get cache item: %w", err)
	}

	return val, nil
}

// GetAndDelete implements Cache.
func (r *redisCache) GetAndDelete(ctx context.Context, key string) (string, error) {
	result, err := r.client.Eval(ctx, getAndDeleteScript, []string{r.key}, key).Result()
	if err != nil {
		return "", fmt.Errorf("can't get cache item: %w", err)
	}

	if value, ok := result.(string); ok {
		return value, nil
	}

	return "", ErrKeyNotFound
}

// Set implements Cache.
func (r *redisCache) Set(ctx context.Context, key string, value string, opts ...Option) error {
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
		return fmt.Errorf("can't set cache item: %w", err)
	}

	return nil
}

// SetOrFail implements Cache.
func (r *redisCache) SetOrFail(ctx context.Context, key string, value string, opts ...Option) error {
	val, err := r.client.HSetNX(ctx, r.key, key, value).Result()
	if err != nil {
		return fmt.Errorf("can't set cache item: %w", err)
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
		if err := r.client.HExpireAt(ctx, r.key, options.validUntil).Err(); err != nil {
			return fmt.Errorf("can't set cache item ttl: %w", err)
		}
	}

	return nil
}

func (r *redisCache) Close() error {
	if r.ownedClient {
		return r.client.Close()
	}

	return nil
}

var _ Cache = (*redisCache)(nil)
