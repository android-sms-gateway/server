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
  for i = 1, #items, 2 do
    redis.call('HDEL', KEYS[1], items[i])
  end
end
return items
`
)

type redisCache struct {
	client *redis.Client

	key string

	ttl time.Duration
}

func NewRedis(client *redis.Client, prefix string, ttl time.Duration) Cache {
	if prefix != "" && !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}

	return &redisCache{
		client: client,

		key: prefix + redisCacheKey,

		ttl: ttl,
	}
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

	if err := r.client.HSet(ctx, r.key, key, value).Err(); err != nil {
		return fmt.Errorf("can't set cache item: %w", err)
	}

	if !options.validUntil.IsZero() {
		if err := r.client.HExpireAt(ctx, r.key, options.validUntil, key).Err(); err != nil {
			return fmt.Errorf("can't set cache item ttl: %w", err)
		}
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
