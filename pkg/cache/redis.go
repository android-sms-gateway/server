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

	hgetallAndDeleteScript = `
local items = redis.call('HGETALL', KEYS[1])
if #items > 0 then
	 local ok = pcall(redis.call, 'UNLINK', KEYS[1])
	 if not ok then redis.call('DEL', KEYS[1]) end
end
return items
`

	// getAndUpdateTTLScript atomically gets a hash field and updates its TTL
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
  local newTtl = ttl + ttlDelta
  redis.call('HExpire', KEYS[1], newTtl, field)
end

return value
`
)

type redisCache struct {
	client *redis.Client

	key string

	ttl time.Duration
}

func NewRedis(client *redis.Client, prefix string, ttl time.Duration) *redisCache {
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
func (r *redisCache) Drain(ctx context.Context) (map[string][]byte, error) {
	res, err := r.client.Eval(ctx, hgetallAndDeleteScript, []string{r.key}).Result()
	if err != nil {
		return nil, fmt.Errorf("can't drain cache: %w", err)
	}

	arr, ok := res.([]any)
	if !ok || len(arr) == 0 {
		return map[string][]byte{}, nil
	}

	out := make(map[string][]byte, len(arr)/2)
	for i := 0; i < len(arr); i += 2 {
		f, _ := arr[i].(string)
		v, _ := arr[i+1].(string)
		out[f] = []byte(v)
	}

	return out, nil
}

// Get implements Cache.
func (r *redisCache) Get(ctx context.Context, key string, opts ...GetOption) ([]byte, error) {
	o := getOptions{}
	o.apply(opts...)

	if o.isEmpty() {
		// No options, simple get
		val, err := r.client.HGet(ctx, r.key, key).Result()
		if err != nil {
			if err == redis.Nil {
				return nil, ErrKeyNotFound
			}

			return nil, fmt.Errorf("can't get cache item: %w", err)
		}

		return []byte(val), nil
	}

	// Handle TTL options atomically using Lua script
	var ttlTimestamp, ttlDelta int64
	if o.validUntil != nil {
		ttlTimestamp = o.validUntil.Unix()
	} else if o.setTTL != nil {
		ttlTimestamp = time.Now().Add(*o.setTTL).Unix()
	} else if o.updateTTL != nil {
		ttlDelta = int64(o.updateTTL.Seconds())
	} else if o.defaultTTL {
		ttlTimestamp = time.Now().Add(r.ttl).Unix()
	} else {
		// No TTL options, fallback to simple get
		val, err := r.client.HGet(ctx, r.key, key).Result()
		if err != nil {
			if err == redis.Nil {
				return nil, ErrKeyNotFound
			}
			return nil, fmt.Errorf("can't get cache item: %w", err)
		}
		return []byte(val), nil
	}

	delArg := "0"
	if o.delete {
		delArg = "1"
	}

	// Use atomic get and TTL update script
	result, err := r.client.Eval(ctx, getAndUpdateTTLScript, []string{r.key}, key, delArg, ttlTimestamp, ttlDelta).Result()
	if err != nil {
		return nil, fmt.Errorf("can't get cache item: %w", err)
	}

	if value, ok := result.(string); ok {
		return []byte(value), nil
	}

	return nil, ErrKeyNotFound
}

// GetAndDelete implements Cache.
func (r *redisCache) GetAndDelete(ctx context.Context, key string) ([]byte, error) {
	return r.Get(ctx, key, AndDelete())
}

// Set implements Cache.
func (r *redisCache) Set(ctx context.Context, key string, value []byte, opts ...Option) error {
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
func (r *redisCache) SetOrFail(ctx context.Context, key string, value []byte, opts ...Option) error {
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
		if err := r.client.HExpireAt(ctx, r.key, options.validUntil, key).Err(); err != nil {
			return fmt.Errorf("can't set cache item ttl: %w", err)
		}
	}

	return nil
}

var _ Cache = (*redisCache)(nil)
