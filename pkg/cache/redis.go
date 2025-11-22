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
  redis.call('HExpireAt', KEYS[1], ttlTs, 'FIELDS', '1', field)
elseif ttlDelta > 0 then
  local ttl = redis.call('HTTL', KEYS[1], 'FIELDS', '1', field)
  if ttl < 0 then
    ttl = 0
  end
  local newTtl = ttl + ttlDelta
  redis.call('HExpire', KEYS[1], newTtl, 'FIELDS', '1', field)
end

return value
`
)

// RedisConfig configures the Redis cache backend.
//
// This struct provides configuration options for creating a Redis-based cache
// implementation. You can either provide an existing Redis client or let the
// cache create one from a URL.
type RedisConfig struct {
	// Client is the Redis client to use.
	// If nil, a client is created from the URL.
	Client *redis.Client

	// URL is the Redis URL to use.
	// If empty, the Redis client is not created.
	URL string

	// Prefix is the prefix to use for all keys in the Redis cache.
	// This helps avoid key collisions when multiple applications use the same Redis instance.
	Prefix string

	// TTL is the time-to-live for all cache entries.
	// This is the default TTL used when no explicit TTL is provided.
	TTL time.Duration
}

// RedisCache implements the Cache interface using Redis as the backend.
//
// This implementation stores all data in a Redis hash, with each cache item
// being a field in the hash. It uses Redis's built-in TTL functionality for
// expiration and Lua scripts for atomic operations.
//
// The Redis cache is suitable for:
//   - Distributed applications where multiple processes need access to the same cache
//   - Caching large amounts of data that don't fit in memory
//   - Scenarios where cache persistence is required
//   - High-availability caching with Redis clustering
//
// For single-process applications or when low latency is critical, consider using
// the in-memory implementation instead.
type RedisCache struct {
	client      *redis.Client
	ownedClient bool

	key string

	ttl time.Duration
}

// NewRedis creates a new Redis cache with the specified configuration.
//
// This function validates the configuration and creates a Redis client if one
// is not provided. The key used for storing cache items in Redis is constructed
// from the prefix and a constant suffix.
//
// Parameters:
//   - config: Configuration for the Redis cache
//
// Returns:
//   - *redisCache: A new Redis cache instance
//   - error: An error if the configuration is invalid or the Redis client cannot be created
//
// Example:
//
//	// Create a Redis cache with a new client
//	config := cache.RedisConfig{
//	    URL:    "redis://localhost:6379",
//	    Prefix: "myapp:",
//	    TTL:    time.Hour,
//	}
//	redisCache, err := cache.NewRedis(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer redisCache.Close()
//
//	// Create a Redis cache with an existing client
//	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
//	config := cache.RedisConfig{
//	    Client: client,
//	    Prefix: "myapp:",
//	    TTL:    time.Hour,
//	}
//	redisCache, err := cache.NewRedis(config)
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

// Cleanup removes all expired items from the Redis cache.
//
// For Redis cache implementation, this is a no-op because Redis handles
// expiration automatically. Expired items are automatically removed by Redis
// based on their TTL settings.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (currently unused but kept for interface compatibility)
//
// Returns:
//   - error: Always nil for Redis cache
//
// Example:
//
//	// No-op for Redis cache, but included for interface compatibility
//	err := redisCache.Cleanup(ctx)
func (r *RedisCache) Cleanup(_ context.Context) error {
	return nil
}

// Delete removes the item associated with the given key from the Redis cache.
//
// If the key does not exist, this operation performs no action and returns nil.
// The operation uses Redis's HDEL command to remove the field from the hash.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: The key to remove from the cache
//
// Returns:
//   - error: nil on success, or an error if the Redis operation fails
//
// Example:
//
//	// Remove a specific key
//	err := redisCache.Delete(ctx, "user:123")
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.HDel(ctx, r.key, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache item: %w", err)
	}

	return nil
}

// Drain returns a map of all non-expired items in the Redis cache and clears the cache.
//
// This operation uses a Lua script to atomically get all fields from the hash
// and then delete the entire hash. This ensures that the operation is atomic
// and no items are lost between the get and delete operations.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns:
//   - map[string][]byte: A map containing all non-expired key-value pairs
//   - error: nil on success, or an error if the Redis operation fails
//
// Example:
//
//	// Get all items and clear the cache
//	items, err := redisCache.Drain(ctx)
//	for key, value := range items {
//	    log.Printf("Drained: %s = %s", key, string(value))
//	}
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

// Get retrieves the value for the given key from the Redis cache.
//
// The behavior depends on the key's existence and expiration state:
//   - If the key exists and has not expired, returns the value and nil error
//   - If the key does not exist, returns nil and ErrKeyNotFound
//   - If the key exists but has expired, returns nil and ErrKeyExpired
//
// GetOptions can be used to modify the behavior, such as updating TTL,
// deleting the key after retrieval, or setting a new expiration time.
// When GetOptions are provided, a Lua script is used for atomic operations.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: The key to retrieve
//   - opts: Optional operations to perform during retrieval (e.g., AndDelete, AndSetTTL)
//
// Returns:
//   - []byte: The cached value if found and not expired
//   - error: nil on success, ErrKeyNotFound if key doesn't exist,
//     ErrKeyExpired if key exists but has expired, otherwise an error
//
// Example:
//
//	// Simple get
//	value, err := redisCache.Get(ctx, "user:123")
//
//	// Get and extend TTL by 30 minutes
//	value, err := redisCache.Get(ctx, "session:abc", cache.AndUpdateTTL(30*time.Minute))
//
//	// Get and delete atomically
//	value, err := redisCache.Get(ctx, "temp:xyz", cache.AndDelete())
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

// GetAndDelete retrieves the value for the given key and atomically deletes it from the Redis cache.
//
// This is equivalent to calling Get with the AndDelete option, but provides a more
// convenient API for the common pattern of reading and removing a value in one operation.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: The key to retrieve and delete
//
// Returns:
//   - []byte: The cached value if found and not expired
//   - error: nil on success, ErrKeyNotFound if key doesn't exist,
//     ErrKeyExpired if key exists but has expired, otherwise an error
//
// Example:
//
//	// Atomically get and remove a value
//	value, err := redisCache.GetAndDelete(ctx, "queue:item:123")
func (r *RedisCache) GetAndDelete(ctx context.Context, key string) ([]byte, error) {
	return r.Get(ctx, key, AndDelete())
}

// Set stores the value for the given key in the Redis cache, overwriting any existing value.
//
// The value will be stored with the default TTL configured for the cache,
// unless overridden by options. If the key already exists, its value and TTL will be updated.
// This method uses Redis pipelining to ensure that the set and TTL operations are atomic.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: The key to store the value under
//   - value: The value to store as a byte slice
//   - opts: Optional configuration for this specific item (e.g., custom TTL)
//
// Returns:
//   - error: nil on success, or an error if the Redis operation fails
//
// Example:
//
//	// Set with default TTL
//	err := redisCache.Set(ctx, "user:123", []byte("user data"))
//
//	// Set with custom TTL
//	err := redisCache.Set(ctx, "session:abc", []byte("session data"), cache.WithTTL(30*time.Minute))
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

// SetOrFail stores the value for the given key only if the key does not already exist.
//
// This is an atomic operation that prevents race conditions when multiple goroutines
// might try to set the same key simultaneously. It uses Redis's HSetNX command
// which only sets the field if it doesn't already exist.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - key: The key to store the value under
//   - value: The value to store as a byte slice
//   - opts: Optional configuration for this specific item (e.g., custom TTL)
//
// Returns:
//   - error: nil on success, ErrKeyExists if the key already exists and is not expired,
//     otherwise an error
//
// Example:
//
//	// Try to set a value only if key doesn't exist
//	err := redisCache.SetOrFail(ctx, "lock:resource", []byte("locked"))
//	if errors.Is(err, cache.ErrKeyExists) {
//	    // Key already exists, handle conflict
//	}
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

// Close releases any resources held by the Redis cache.
//
// If the Redis cache was created with a client (rather than using an existing client),
// this method will close the Redis client connection. If an existing client was provided,
// this method is a no-op to avoid closing a client that might be used elsewhere.
//
// Returns:
//   - error: nil on success, or an error if closing the Redis client fails
//
// Example:
//
//	// Properly close the cache when done
//	defer redisCache.Close()
func (r *RedisCache) Close() error {
	if r.ownedClient {
		if err := r.client.Close(); err != nil {
			return fmt.Errorf("failed to close redis client: %w", err)
		}
	}

	return nil
}

// Compile-time check to ensure redisCache implements the Cache interface.
var _ Cache = (*RedisCache)(nil)
